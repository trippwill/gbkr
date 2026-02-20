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
	sess, err := gbkr.Session(client)
	if err != nil {
		log.Fatalf("Session capability denied: %v", err)
	}

	session, err := sess.InitBrokerageSession(ctx, &models.SSOInitRequest{})
	if err != nil {
		log.Fatalf("Error initializing SSO: %v", err)
	}
	fmt.Printf("Session Authenticated: %v\n", session.Authenticated)
	fmt.Printf("Session Connected: %v\n", session.Connected)

	status, err := sess.SessionStatus(ctx)
	if err != nil {
		log.Fatalf("Error getting auth status: %v", err)
	}
	fmt.Printf("Auth Status: %v\n", status.Authenticated)

	// Accounts: discover available accounts
	lister, err := gbkr.Accounts(client)
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
	reader, err := gbkr.Account(client, acctID)
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
	md, err := gbkr.MarketData(client)
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

	fmt.Println("\nDone.")
}
