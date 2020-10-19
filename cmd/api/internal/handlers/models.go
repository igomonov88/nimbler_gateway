// models.go contains all needed structs corresponding to API Contract.

package handler

// CreateUserRequest create user with given email and password.
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUserResponse returns user_id of successfully created user
type CreateUserResponse struct {
	UserID string `json:"user_id"`
}

// RetrieveUserResponse return info about user by given user_id
type RetrieveUserResponse struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// UpdateUserInfoRequest needs to update user info
type UpdateUserInfoRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// UpdateUserPasswordRequest needs to update user's password
type UpdateUserPasswordRequest struct {
	UserID string `json:"user_id"`
	Password string `json:"password"`
}

// DoesEmailExistsRequest needs to know what such email is already exist
type DoesEmailExistsRequest struct {
	Email string `json:"email"`
}

// DoesEmailExistsResponse returns email is already exist
type DoesEmailExistsResponse struct {
	Exist bool `json:"exist"`
}