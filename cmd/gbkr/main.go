package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/models"
)

func main() {
	baseURL := flag.String("base-url", "https://localhost:5000/v1/api", "IBKR API base URL")
	permsFile := flag.String("permissions-file", "", "YAML permissions file (optional floor; JIT prompts for anything missing)")
	insecure := flag.Bool("insecure", false, "Skip TLS verification")
	flag.Parse()

	opts := []gbkr.Option{
		gbkr.WithBaseURL(*baseURL),
	}
	if *insecure {
		opts = append(opts, gbkr.WithInsecureTLS())
	}
	if *permsFile != "" {
		opts = append(opts, gbkr.WithPermissionsFromFile(*permsFile))
	}
	opts = append(opts, gbkr.WithInteractivePrompt())

	client, err := gbkr.NewClient(opts...)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	ctx := context.Background()

	// Session: elevate to brokerage session
	bc, err := client.BrokerageSession(ctx, &models.SSOInitRequest{})
	if err != nil {
		log.Fatalf("Brokerage session denied: %v", err)
	}

	status, err := client.SessionStatus(ctx)
	if err != nil {
		log.Fatalf("Error getting auth status: %v", err)
	}
	fmt.Printf("Session Authenticated: %v\n", status.Authenticated)
	fmt.Printf("Session Connected: %v\n", status.Connected)

	fmt.Printf("Auth Status: %v\n", status.Authenticated)

	// Accounts: discover available accounts
	lister, err := bc.Accounts()
	if err != nil {
		if errors.Is(err, gbkr.ErrPermissionDenied) {
			log.Fatalf("Accounts capability denied: %v", err)
		}
		log.Fatalf("Error creating Accounts capability: %v", err)
	}

	accountList, err := lister.ListAccounts(ctx)
	if err != nil {
		log.Fatalf("Error listing accounts: %v", err)
	}
	if len(accountList.Accounts) == 0 {
		log.Fatal("No accounts found")
	}
	for _, id := range accountList.Accounts {
		fmt.Printf("Account: %s\n", id)
	}

	// Scope to the first account
	acctID := accountList.Accounts[0]
	reader, err := bc.Account(acctID)
	if err != nil {
		if errors.Is(err, gbkr.ErrPermissionDenied) {
			log.Fatalf("Account reader denied: %v", err)
		}
		log.Fatalf("Error creating account reader: %v", err)
	}
	fmt.Printf("\nScoped to account: %s\n", reader.AccountID())

	summary, err := reader.Summary(ctx)
	if err != nil {
		log.Printf("Error getting summary: %v", err)
	} else {
		fmt.Printf("Account ID: %s\n", summary.AccountID)
	}

	// Positions: obtained from the scoped account reader
	positions, err := reader.Positions()
	if err != nil {
		if errors.Is(err, gbkr.ErrPermissionDenied) {
			log.Fatalf("Positions capability denied: %v", err)
		}
		log.Fatalf("Error creating positions reader: %v", err)
	}

	var conIDs []models.ConID
	pos, err := positions.ListPositions(ctx, 0)
	if err != nil {
		log.Printf("Error getting positions: %v", err)
	} else {
		for _, p := range pos {
			fmt.Printf("Position: %s qty=%.0f\n", p.ContractDesc, p.Qty)
			if p.AssetClass != models.AssetOption && p.ConID != 0 {
				conIDs = append(conIDs, p.ConID)
			}
		}
	}

	// Market data
	md, err := bc.MarketData()
	if err != nil {
		if errors.Is(err, gbkr.ErrPermissionDenied) {
			log.Fatalf("MarketData capability denied: %v", err)
		}
		log.Fatalf("Error creating MarketData capability: %v", err)
	}
	snapshots, err := md.Snapshot(ctx, models.SnapshotParams{
		ConIDs: conIDs,
		Fields: models.FieldsQuote,
	})
	if err != nil {
		log.Printf("Error getting market data snapshots: %v", err)
		return
	}
	for _, snap := range snapshots {
		fmt.Printf(
			"Market Data(%v): %s last=%s bid=%s ask=%s vol=%s\n",
			snap.UpdateTime,
			snap.Get(models.FieldSymbol),
			snap.Get(models.FieldLast),
			snap.Get(models.FieldBid),
			snap.Get(models.FieldAsk),
			snap.Get(models.FieldVolume),
		)
	}

	// Contracts: look up contract details
	contracts, err := bc.Contracts()
	if err != nil {
		if errors.Is(err, gbkr.ErrPermissionDenied) {
			log.Printf("Contracts capability denied: %v", err)
		} else {
			log.Printf("Error creating Contracts capability: %v", err)
		}
	} else if len(conIDs) > 0 {
		details, err := contracts.Info(ctx, conIDs[0])
		if err != nil {
			log.Printf("Error getting contract info: %v", err)
		} else {
			fmt.Printf("Contract: conid=%d symbol=%s\n", details.ConID, details.Symbol)
		}
	}

	// Trades: recent trade executions (brokerage session)
	trades, err := bc.Trades()
	if err != nil {
		if errors.Is(err, gbkr.ErrPermissionDenied) {
			log.Printf("Trades capability denied: %v", err)
		} else {
			log.Printf("Error creating Trades capability: %v", err)
		}
	} else {
		recent, err := trades.RecentTrades(ctx, 1)
		if err != nil {
			log.Printf("Error getting recent trades: %v", err)
		} else {
			fmt.Printf("Recent trades: %d\n", len(recent))
		}
	}

	// Transaction history: read-only (no brokerage session required)
	txn, err := client.TransactionHistory()
	if err != nil {
		if errors.Is(err, gbkr.ErrPermissionDenied) {
			log.Printf("TransactionHistory capability denied: %v", err)
		} else {
			log.Printf("Error creating TransactionHistory capability: %v", err)
		}
	} else if len(accountList.Accounts) > 0 && len(conIDs) > 0 {
		hist, err := txn.TransactionHistory(ctx, accountList.Accounts[0], conIDs[0], 7)
		if err != nil {
			log.Printf("Error getting transaction history: %v", err)
		} else {
			fmt.Printf("Transactions: %d\n", len(hist.Transactions))
		}
	}

	fmt.Println("\nDone.")
}
