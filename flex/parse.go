package flex

import (
	"encoding/xml"
	"fmt"
	"io"
)

// ParseActivityStatement decodes an Activity Flex Query XML response from r
// into a [QueryResponse]. Missing optional sections (Trades, CashTransactions,
// etc.) produce empty slices rather than errors.
//
// The response must contain at least one FlexStatement with a non-empty AccountID.
func ParseActivityStatement(r io.Reader) (*QueryResponse, error) {
	var resp QueryResponse
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode flex response: %w", err)
	}

	if len(resp.Statements) == 0 {
		return nil, fmt.Errorf("flex response contains no statements")
	}

	for i := range resp.Statements {
		if resp.Statements[i].AccountID == "" {
			return nil, fmt.Errorf("statement %d: missing accountId", i)
		}
	}

	return &resp, nil
}

// parseSendRequestResponse decodes a SendRequest endpoint XML response.
func parseSendRequestResponse(r io.Reader) (*sendRequestResponse, error) {
	var resp sendRequestResponse
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode send-request response: %w", err)
	}
	return &resp, nil
}
