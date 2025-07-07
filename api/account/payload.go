package account

// PayloadAccount represents the payload for creating an account
type PayloadAccount struct {
	Name    string `json:"name"`
	Type    Type   `json:"type"`
	Balance int64  `json:"balance"`
}
