package models

type DBStats struct {
	Open         int  `json:"open"`
	Iddle        int  `json:"iddle"`
	OpenTimeout  int  `json:"openTimeout"`
	IddleTimeout int  `json:"iddleTimeout"`
	Reconnect    bool `json:"reconnect"`
}
