package gbkr_test

import (
	"fmt"
	"log"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/models"
)

func ExampleNewClient() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithInsecureTLS(),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = client
	fmt.Println("client created")
	// Output: client created
}

func ExampleClient_SessionStatus() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client // use client.SessionStatus(ctx)
	fmt.Println("client for session status created")
	// Output: client for session status created
}

func ExampleBrokerageClient_Accounts() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	accts := bc.Accounts()
	_ = accts // use accts.List() or accts.PnL()
	fmt.Println("account lister created")
	// Output: account lister created
}

func ExampleBrokerageClient_Account() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	reader := bc.Account(models.AccountID("U1234567"))
	fmt.Println("account reader for:", reader.ID())
	// Output: account reader for: U1234567
}

func ExampleClient_Portfolio() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}

	pr := client.Portfolio(models.AccountID("U1234567"))
	fmt.Println("portfolio reader for:", pr.ID())
	// Output: portfolio reader for: U1234567
}

func ExampleClient_Analysis() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ar := client.Analysis()
	_ = ar // use ar.Transactions(ctx, accountID, conID, days)
	fmt.Println("analysis reader created")
	// Output: analysis reader created
}

func ExampleBrokerageClient_MarketData() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	md := bc.MarketData()
	_ = md // use md.Snapshot() or md.History()
	fmt.Println("market data reader created")
	// Output: market data reader created
}

func ExampleBrokerageClient_Contracts() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	cr := bc.Contracts()
	_ = cr // use cr.Info()
	fmt.Println("contract reader created")
	// Output: contract reader created
}

func ExampleBrokerageClient_SecurityDefinitions() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	sd := bc.SecurityDefinitions()
	_ = sd // use sd.Search()
	fmt.Println("security definition reader created")
	// Output: security definition reader created
}

func ExampleBrokerageClient_Trades() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// In production, use client.BrokerageSession(ctx, req) to obtain a BrokerageClient.
	// Direct construction is used here because examples don't make real HTTP calls.
	bc := &gbkr.BrokerageClient{Client: client}

	tr := bc.Trades()
	_ = tr // use tr.Recent()
	fmt.Println("trade reader created")
	// Output: trade reader created
}

func ExampleClient_BrokerageSession() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client // use client.BrokerageSession(ctx, req) to elevate
	fmt.Println("client ready for elevation")
	// Output: client ready for elevation
}
