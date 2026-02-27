package gbkr

import (
	"context"
	"encoding/json"
	"time"

	"github.com/trippwill/gbkr/internal/jx"
)

// SessionStatus is the response for POST /iserver/auth/ssodh/init and POST /iserver/auth/status.
type SessionStatus struct {
	// Authenticated indicates if the session is authenticated.
	Authenticated bool
	// Connected indicates if connected to the brokerage.
	Connected bool
	// Competing indicates if another session is competing.
	Competing bool
	// Established indicates if the session is established.
	Established bool
	// Fail contains the failure message; empty on success.
	Fail string
	// MAC is the MAC address.
	MAC string
	// HardwareInfo is the hardware info. (API: "hardware_info")
	HardwareInfo string
	// ServerName is the server name.
	ServerName string
	// ServerVersion is the server version.
	ServerVersion string
}

func (s *SessionStatus) UnmarshalJSON(data []byte) error {
	var raw struct {
		Authenticated *bool   `json:"authenticated,omitempty"`
		Connected     *bool   `json:"connected,omitempty"`
		Competing     *bool   `json:"competing,omitempty"`
		Established   *bool   `json:"established,omitempty"`
		Fail          *string `json:"fail,omitempty"`
		MAC           *string `json:"MAC,omitempty"`
		HardwareInfo  *string `json:"hardware_info,omitempty"`
		ServerName    *string `json:"serverName,omitempty"`
		ServerVersion *string `json:"serverVersion,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Authenticated = jx.Deref(raw.Authenticated)
	s.Connected = jx.Deref(raw.Connected)
	s.Competing = jx.Deref(raw.Competing)
	s.Established = jx.Deref(raw.Established)
	s.Fail = jx.Deref(raw.Fail)
	s.MAC = jx.Deref(raw.MAC)
	s.HardwareInfo = jx.Deref(raw.HardwareInfo)
	s.ServerName = jx.Deref(raw.ServerName)
	s.ServerVersion = jx.Deref(raw.ServerVersion)
	return nil
}

// SessionStatus checks the current brokerage session status
// (POST /iserver/auth/status).
func (c *Client) SessionStatus(ctx context.Context) (*SessionStatus, error) {
	start := time.Now()
	var result SessionStatus
	err := c.doPost(ctx, "/iserver/auth/status", nil, &result)
	c.emitOp(ctx, OpSessionStatus, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
