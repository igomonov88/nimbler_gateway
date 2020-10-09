// models.go contains all needed structs corresponding to API Contract.

package handler

// CreateUserRequest create user with given email and password.
type CreateUserRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

// CreateUserResponse returns user_id of successfully created user
type CreateUserResponse struct {
	UserID string `json:"user_id"`
}
