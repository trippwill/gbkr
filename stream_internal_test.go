package gbkr

import (
	"testing"
)

func TestDeriveWSURL(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		want    string
		wantErr bool
	}{
		{"https", "https://localhost:5000/v1/api", "wss://localhost:5000/v1/api/ws", false},
		{"http", "http://localhost:5000/v1/api", "ws://localhost:5000/v1/api/ws", false},
		{"bad scheme", "ftp://localhost/api", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deriveWSURL(tt.base)
			if (err != nil) != tt.wantErr {
				t.Fatalf("deriveWSURL(%q) error = %v, wantErr %v", tt.base, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("deriveWSURL(%q) = %q, want %q", tt.base, got, tt.want)
			}
		})
	}
}
