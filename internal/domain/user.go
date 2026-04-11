package domain

// Role represents a user's access level.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleUploader Role = "uploader"
)

// User represents an authenticated system user.
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         Role
}
