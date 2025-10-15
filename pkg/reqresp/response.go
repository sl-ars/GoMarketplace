package reqresp

type StandardResponse struct {
	Success bool        `json:"success" example:"true" extensions:"x-order=1"`
	Message string      `json:"message" example:"Operation completed successfully" extensions:"x-order=2"`
	Data    interface{} `json:"data,omitempty" extensions:"x-order=3"`
	Error   *ErrorInfo  `json:"error,omitempty" extensions:"x-order=4"`
}

type ErrorInfo struct {
	Code    int    `json:"code" example:"400" extensions:"x-order=1"`
	Details string `json:"details" example:"Invalid ID format: expected integer, got 'abc'" extensions:"x-order=2"`
}
