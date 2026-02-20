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

func ExampleSession() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnlyAuth()),
	)
	if err != nil {
		log.Fatal(err)
	}

	sc, err := gbkr.Session(client)
	if err != nil {
		log.Fatal(err)
	}
	_ = sc // use sc.InitBrokerageSession() or sc.SessionStatus()
	fmt.Println("session client created")
	// Output: session client created
}

func ExampleAccounts() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnly()),
	)
	if err != nil {
		log.Fatal(err)
	}

	accts, err := gbkr.Accounts(client)
	if err != nil {
		log.Fatal(err)
	}
	_ = accts // use accts.ListAccounts() or accts.AccountPnL()
	fmt.Println("account lister created")
	// Output: account lister created
}

func ExampleAccount() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnly()),
	)
	if err != nil {
		log.Fatal(err)
	}

	reader, err := gbkr.Account(client, models.AccountID("U1234567"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("account reader for:", reader.AccountID())
	// Output: account reader for: U1234567
}

func ExamplePositions() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}

	pos, err := gbkr.Positions(client, models.AccountID("U1234567"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("position reader for:", pos.AccountID())
	// Output: position reader for: U1234567
}

func ExampleMarketData() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.ReadOnly()),
	)
	if err != nil {
		log.Fatal(err)
	}

	md, err := gbkr.MarketData(client)
	if err != nil {
		log.Fatal(err)
	}
	_ = md // use md.Snapshot() or md.History()
	fmt.Println("market data reader created")
	// Output: market data reader created
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

// ExampleAccount_positions demonstrates getting a PositionReader
// from an AccountReader, which checks portfolio permissions at that point.
func ExampleAccount_positions() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
		gbkr.WithPermissions(gbkr.AllPermissions()),
	)
	if err != nil {
		log.Fatal(err)
	}

	acct, err := gbkr.Account(client, models.AccountID("U1234567"))
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
