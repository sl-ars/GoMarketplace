package apperror

// Common API error messages - safe to return to clients
// Never expose internal Go error details!
const (
	// Generic errors
	ErrInternalServer   = "An internal error occurred. Please try again later."
	ErrBadRequest       = "Invalid request. Please check your input."
	ErrUnauthorized     = "Authentication required."
	ErrForbidden        = "You do not have permission to perform this action."
	ErrNotFound         = "The requested resource was not found."
	ErrConflict         = "Resource already exists."
	ErrValidationFailed = "Validation failed. Please check your input."
	ErrInvalidJSON      = "Invalid JSON format."
	ErrInvalidID        = "Invalid ID format."
	ErrTooManyRequests  = "Rate limit exceeded. Please slow down and try again later."

	// Auth errors
	ErrInvalidCredentials   = "Invalid email or password."
	ErrInvalidToken         = "Invalid or expired token."
	ErrMissingToken         = "Authentication token is required."
	ErrEmailTaken           = "This email is already registered."
	ErrUsernameTaken        = "This username is already taken."
	ErrUserNotFound         = "User not found."
	ErrEmailNotVerified     = "Please verify your email address before logging in."
	ErrEmailAlreadyVerified = "Email is already verified."
	ErrPasswordResetFailed  = "Password reset failed. Please try again."

	// Cart errors
	ErrCartEmpty        = "Your cart is empty."
	ErrCartItemNotFound = "Cart item not found."
	ErrAddToCartFailed  = "Failed to add item to cart."
	ErrRemoveFromCart   = "Failed to remove item from cart."
	ErrClearCartFailed  = "Failed to clear cart."
	ErrCartFetchFailed  = "Failed to retrieve cart."

	// Product errors
	ErrProductNotFound     = "Product not found."
	ErrProductCreateFailed = "Failed to create product."
	ErrProductFetchFailed  = "Failed to retrieve products."

	// Offer errors
	ErrOfferNotFound     = "Offer not found."
	ErrOfferCreateFailed = "Failed to create offer."
	ErrOfferUpdateFailed = "Failed to update offer."
	ErrOfferDeleteFailed = "Failed to delete offer."
	ErrOfferFetchFailed  = "Failed to retrieve offers."
	ErrNotOfferOwner     = "You do not have permission to modify this offer."

	// Order errors
	ErrOrderNotFound      = "Order not found."
	ErrOrderCreateFailed  = "Failed to create order."
	ErrOrderCancelFailed  = "Failed to cancel order."
	ErrOrderFetchFailed   = "Failed to retrieve orders."
	ErrCheckoutFailed     = "Checkout failed. Please try again."
	ErrPaymentFailed      = "Payment processing failed."
	ErrOrderAccessDenied  = "You do not have access to this order."
	ErrInvalidOrderStatus = "Invalid order status."

	// Refund errors
	ErrRefundNotFound     = "Refund request not found."
	ErrRefundCreateFailed = "Failed to create refund request."
	ErrRefundUpdateFailed = "Failed to update refund status."
	ErrRefundNotAllowed   = "Refund is not allowed for this item."
)

// AppError represents an application error with a safe message for clients
type AppError struct {
	// Message is safe to return to clients
	Message string
	// Err is the underlying error (for logging only, never expose!)
	Err error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// New creates a new AppError with a safe message
func New(message string, err error) *AppError {
	return &AppError{
		Message: message,
		Err:     err,
	}
}

// SafeMessage returns the client-safe error message
func (e *AppError) SafeMessage() string {
	return e.Message
}
