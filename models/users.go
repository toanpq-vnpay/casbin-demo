package models

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// Create extended response with user info and groups
type UserResponse struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Groups   []string `json:"groups"`
}
