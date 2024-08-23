package entities

// User represents a user returned by user.query.
type User struct {
	PHID     string   `json:"phid"`
	UserName string   `json:"userName"`
	RealName string   `json:"realName"`
	Email    string   `json:"email"`
	Image    string   `json:"image"`
	URI      string   `json:"uri"`
	Roles    []string `json:"roles"`
}
