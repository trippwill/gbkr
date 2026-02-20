package models

import "encoding/json"

// SSOInitRequest is the request body for POST /iserver/auth/ssodh/init.
type SSOInitRequest struct {
	// Compete determines if other brokerage sessions should be disconnected.
	Compete *bool `json:"compete,omitempty"`

	// Publish publishes the brokerage session token when the session is initialized.
	Publish *bool `json:"publish,omitempty"`
}

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
	s.Authenticated = deref(raw.Authenticated)
	s.Connected = deref(raw.Connected)
	s.Competing = deref(raw.Competing)
	s.Established = deref(raw.Established)
	s.Fail = deref(raw.Fail)
	s.MAC = deref(raw.MAC)
	s.HardwareInfo = deref(raw.HardwareInfo)
	s.ServerName = deref(raw.ServerName)
	s.ServerVersion = deref(raw.ServerVersion)
	return nil
}
