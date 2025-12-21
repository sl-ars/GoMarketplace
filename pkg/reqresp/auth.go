package reqresp

// RegisterRequest represents user registration input
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterResponse represents successful registration output
type RegisterResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// LoginRequest represents user login input
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents successful login output with tokens
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshRequest represents token refresh input
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents token refresh output
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

// VerifyResponse represents token verification output
type VerifyResponse struct {
	Valid  bool  `json:"valid"`
	UserID int64 `json:"user_id,omitempty"`
}
