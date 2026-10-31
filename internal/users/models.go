package users

// the business representation of a user
type User struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	RoleID     int64  `json:"role_id"`
	IsVerified bool   `json:"is_verified"`
}
