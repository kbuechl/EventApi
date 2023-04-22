package auth

type UserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Verified  bool   `json:"verified_email"`
	Picture   string `json:"picture"`
	LastName  string `json:"family_name"`
	FirstName string `json:"given_name"`
}
