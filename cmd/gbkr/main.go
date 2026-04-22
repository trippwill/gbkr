package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/brokerage"
)

func main() {
	baseURL := flag.String("base-url", "https://localhost:5000/v1/api", "IBKR API base URL")
	insecure := flag.Bool("insecure", false, "Skip TLS verification")
	stream := flag.Bool("stream", false, "Run streaming demo after REST demo")
	streamDuration := flag.Duration("stream-duration", 30*time.Second, "How long to listen for stream updates")
	flag.Parse()

	opts := []gbkr.Option{
		gbkr.WithBaseURL(*baseURL),
	}
	if *insecure {
		opts = append(opts, gbkr.WithInsecureTLS())
	}

	client, err := gbkr.NewClient(opts...)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	ctx := context.Background()

	// Session: elevate to brokerage session
	bc, err := brokerage.NewSession(ctx, client, &brokerage.SSOInitRequest{})
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
	lister := bc.Accounts()

	accountList, err := lister.List(ctx)
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
	reader := bc.Account(acctID)
	fmt.Printf("\nScoped to account: %s\n", reader.ID())

	summary, err := reader.Summary(ctx)
	if err != nil {
		log.Printf("Error getting summary: %v", err)
	} else {
		fmt.Printf("Account ID: %s\n", summary.AccountID)
	}

	// Positions: obtained from Portfolio (gateway access)
	portfolio := client.Portfolio(gbkr.AccountID(acctID))

	var conIDs []gbkr.ConID
	pos, err := portfolio.Positions(ctx, 0)
	if err != nil {
		log.Printf("Error getting positions: %v", err)
	} else {
		for _, p := range pos {
			fmt.Printf("Position: %s qty=%s\n", p.ContractDesc, p.Qty)
			if p.AssetClass != gbkr.AssetOption && p.ConID != 0 {
				conIDs = append(conIDs, p.ConID)
			}
		}
	}

	// Market data
	md := bc.MarketData()
	snapshots, err := md.Snapshot(ctx, brokerage.SnapshotParams{
		ConIDs: conIDs,
		Fields: brokerage.FieldsQuote,
	})
	if err != nil {
		log.Printf("Error getting market data snapshots: %v", err)
		return
	}
	for _, snap := range snapshots {
		fmt.Printf(
			"Market Data(%v): %s last=%s bid=%s ask=%s vol=%s\n",
			snap.UpdateTime,
			snap.Get(brokerage.FieldSymbol),
			snap.Get(brokerage.FieldLast),
			snap.Get(brokerage.FieldBid),
			snap.Get(brokerage.FieldAsk),
			snap.Get(brokerage.FieldVolume),
		)
	}

	// Contracts: look up contract details
	contracts := bc.Contracts()
	if len(conIDs) > 0 {
		details, err := contracts.Info(ctx, conIDs[0])
		if err != nil {
			log.Printf("Error getting contract info: %v", err)
		} else {
			fmt.Printf("Contract: conid=%d symbol=%s\n", details.ConID, details.Symbol)
		}
	}

	// Trades: recent trade executions (brokerage session)
	trades := bc.Trades()
	recent, err := trades.Recent(ctx, 1)
	if err != nil {
		log.Printf("Error getting recent trades: %v", err)
	} else {
		fmt.Printf("Recent trades: %d\n", len(recent))
	}

	// Analysis: transaction history (gateway access — no brokerage session required)
	analysis := client.Analysis()
	if len(accountList.Accounts) > 0 && len(conIDs) > 0 {
		hist, err := analysis.Transactions(ctx, accountList.Accounts[0], conIDs[0], 7)
		if err != nil {
			log.Printf("Error getting transaction history: %v", err)
		} else {
			fmt.Printf("Transactions: %d\n", len(hist.Value.Transactions))
		}
	}

	// Streaming demo (optional)
	if *stream {
		fmt.Println("\n--- Streaming Demo ---")

		s, err := client.Stream(ctx)
		if err != nil {
			log.Fatalf("Stream: %v", err)
		}
		defer s.Close()

		pnlUpdates, cancelPnL, err := s.PortfolioPnL(gbkr.AccountID(acctID))
		if err != nil {
			log.Printf("PortfolioPnL: %v", err)
		} else {
			defer cancelPnL()
		}

		notifUpdates, cancelNotif, err := s.Notifications()
		if err != nil {
			log.Printf("Notifications: %v", err)
		} else {
			defer cancelNotif()
		}

		var mdUpdates <-chan brokerage.MarketDataUpdate
		var cancelMD func()
		if len(conIDs) > 0 {
			mdUpdates, cancelMD, err = brokerage.SubscribeMarketData(s, conIDs[0], brokerage.FieldLast, brokerage.FieldBid, brokerage.FieldAsk)
			if err != nil {
				log.Printf("SubscribeMarketData: %v", err)
			} else {
				defer cancelMD()
			}
		}

		fmt.Printf("Listening for %v...\n", *streamDuration)

		timer := time.NewTimer(*streamDuration)
		defer timer.Stop()
		for {
			select {
			case u, ok := <-pnlUpdates:
				if !ok {
					pnlUpdates = nil
					continue
				}
				fmt.Printf("[PnL] daily=%.2f unrealized=%.2f realized=%.2f nl=%.2f\n",
					u.DailyPnL, u.UnrealizedPnL, u.RealizedPnL, u.NetLiquidity)
			case n, ok := <-notifUpdates:
				if !ok {
					notifUpdates = nil
					continue
				}
				fmt.Printf("[Notification] %s: %s\n", n.Date, n.Text)
			case u, ok := <-mdUpdates:
				if !ok {
					mdUpdates = nil
					continue
				}
				fmt.Printf("[MarketData] conid=%d fields=%v\n", u.ConID, u.Fields)
			case <-timer.C:
				fmt.Println("Stream duration elapsed.")
				return
			case <-ctx.Done():
				fmt.Println("Context cancelled.")
				return
			}
		}
	}

	fmt.Println("\nDone.")
}
