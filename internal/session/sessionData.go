package session

import (
	"encoding/json"
	"time"
)

type SessionData struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func (sd *SessionData) MarshalBinary() ([]byte, error) {
	return json.Marshal(sd)
}
func (sd *SessionData) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &sd); err != nil {
		return err
	}
	return nil
}
