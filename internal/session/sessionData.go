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
func (sd SessionData) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &sd); err != nil {
		return err
	}
	return nil
}
