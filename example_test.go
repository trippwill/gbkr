package gbkr_test

import (
	"fmt"
	"log"

	"github.com/trippwill/gbkr"
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

func ExampleClient_Portfolio() {
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL("https://localhost:5000/v1/api"),
	)
	if err != nil {
		log.Fatal(err)
	}

	pr := client.Portfolio(gbkr.AccountID("U1234567"))
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
