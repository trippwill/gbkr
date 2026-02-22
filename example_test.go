package gbkr_test

import (
	"context"
	"fmt"
	"log"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/models"
)

func ExampleNewClient() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithInsecureTLS(),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("client created, permissions:", len(client.Permissions()))
	// Output: client created, permissions: 1
}

func ExampleClient_SessionStatus() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnlyAuth()),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client // use client.SessionStatus(ctx) or client.BrokerageSession(ctx, req)
	fmt.Println("client with session permissions created")
	// Output: client with session permissions created
}

func ExampleBrokerageClient_Accounts() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	accts, err := bc.Accounts()
	if err != nil {
		log.Fatal(err)
	}
	_ = accts // use accts.ListAccounts() or accts.AccountPnL()
	fmt.Println("account lister created")
	// Output: account lister created
}

func ExampleBrokerageClient_Account() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	reader, err := bc.Account(models.AccountID("U1234567"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("account reader for:", reader.AccountID())
	// Output: account reader for: U1234567
}

func ExampleClient_Positions() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}

	pos, err := client.Positions(models.AccountID("U1234567"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("position reader for:", pos.AccountID())
	// Output: position reader for: U1234567
}

func ExampleBrokerageClient_MarketData() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	md, err := bc.MarketData()
	if err != nil {
		log.Fatal(err)
	}
	_ = md // use md.Snapshot() or md.History()
	fmt.Println("market data reader created")
	// Output: market data reader created
}

func ExampleBrokerageClient_Contracts() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	cr, err := bc.Contracts()
	if err != nil {
		log.Fatal(err)
	}
	_ = cr // use cr.Info() or cr.Search()
	fmt.Println("contract reader created")
	// Output: contract reader created
}

func ExampleBrokerageClient_Trades() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	tr, err := bc.Trades()
	if err != nil {
		log.Fatal(err)
	}
	_ = tr // use tr.RecentTrades()
	fmt.Println("trade reader created")
	// Output: trade reader created
}

func ExampleClient_BrokerageSession() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnlyAuth()),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client // use client.BrokerageSession(ctx, req) to elevate
	fmt.Println("client ready for elevation")
	// Output: client ready for elevation
}

func ExampleClient_TransactionHistory() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnly()),
	)
	if err != nil {
		log.Fatal(err)
	}

	txn, err := client.TransactionHistory()
	if err != nil {
		log.Fatal(err)
	}
	_ = txn // use txn.TransactionHistory(ctx, accountID, conID, days)
	fmt.Println("transaction reader created")
	// Output: transaction reader created
}

func ExampleNewClient_withPrompter() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnly()),
		gbkr.WithInteractivePrompt(),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = client
	fmt.Println("client with interactive prompter created")
	// Output: client with interactive prompter created
}

func ExampleNewClient_withPermissionsFile() {
	_, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissionsFromFile("permissions.yml"),
	)
	// This would fail without the file, but shows the pattern.
	_ = err
	fmt.Println("permissions file option demonstrated")
	// Output: permissions file option demonstrated
}

// ExampleBrokerageClient_Account_positions demonstrates getting a PositionReader
// from an AccountReader, which checks portfolio permissions at that point.
func ExampleBrokerageClient_Account_positions() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	acct, err := bc.Account(models.AccountID("U1234567"))
	if err != nil {
		log.Fatal(err)
	}

	pos, err := acct.Positions()
	if err != nil {
		log.Fatal(err)
	}

	_ = pos
	ctx := context.Background()
	_ = ctx // use pos.ListPositions(ctx, 0) to fetch positions
	fmt.Println("positions from account reader for:", acct.AccountID())
	// Output: positions from account reader for: U1234567
}
