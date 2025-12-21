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
	Message  string `json:"message,omitempty"`
}

// LoginRequest represents user login input
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents successful login output with tokens
type LoginResponse struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	EmailVerified bool   `json:"email_verified"`
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

// VerifyEmailRequest represents email verification input
type VerifyEmailRequest struct {
	Token string `json:"token,omitempty"` // Token from email link
	Email string `json:"email,omitempty"` // Email + code verification
	Code  string `json:"code,omitempty"`  // 6-digit code
}

// VerifyEmailResponse represents email verification output
type VerifyEmailResponse struct {
	Message string `json:"message"`
}

// ResendVerificationRequest represents resend verification email input
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPasswordRequest represents forgot password input
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPasswordResponse represents forgot password output
type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

// ResetPasswordRequest represents password reset input
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ResetPasswordResponse represents password reset output
type ResetPasswordResponse struct {
	Message string `json:"message"`
}
