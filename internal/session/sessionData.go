package session

import (
	"encoding/json"
)

type SessionData struct {
	Email       string
	AccessToken string
}

func (sd SessionData) MarshalBinary() ([]byte, error) {
	return json.Marshal(sd)
}

// UnmarshalBinary uses a pointer receiver so the unmarshaled data is actually stored.
func (sd *SessionData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, sd)
}
