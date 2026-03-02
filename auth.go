package gbkr

import (
	"context"
	"encoding/json"
	"time"

	"github.com/trippwill/gbkr/internal/jx"
)

// SessionStatus is the response for authentication status endpoints
// (POST /iserver/auth/status, POST /iserver/auth/ssodh/init, POST /iserver/reauthenticate).
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

// Logout ends the current brokerage session (POST /logout).
// Returns an error if the gateway reports an unsuccessful logout.
func (c *Client) Logout(ctx context.Context) error {
	start := time.Now()
	var result struct {
		Status *bool `json:"status"`
	}
	err := c.doPost(ctx, "/logout", nil, &result)
	c.emitOp(ctx, OpLogout, err, time.Since(start))
	if err != nil {
		return err
	}
	if result.Status != nil && !*result.Status {
		return ErrLogoutFailed
	}
	return nil
}

// Reauthenticate triggers re-authentication of the current session
// (POST /iserver/reauthenticate).
func (c *Client) Reauthenticate(ctx context.Context) (*SessionStatus, error) {
	start := time.Now()
	var result SessionStatus
	err := c.doPost(ctx, "/iserver/reauthenticate", nil, &result)
	c.emitOp(ctx, OpReauthenticate, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Tickle sends a keepalive ping to prevent session timeout (POST /tickle).
// Returns the SSO expiry time in milliseconds on success, or an error if the
// gateway reports a failure.
func (c *Client) Tickle(ctx context.Context) (int, error) {
	start := time.Now()
	var result struct {
		SSOExpires int    `json:"ssoExpires"`
		Error      string `json:"error"`
	}
	err := c.doPost(ctx, "/tickle", nil, &result)
	c.emitOp(ctx, OpTickle, err, time.Since(start))
	if err != nil {
		return 0, err
	}
	if result.Error != "" {
		return 0, TickleError(result.Error)
	}
	return result.SSOExpires, nil
}
