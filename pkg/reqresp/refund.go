package reqresp

type RefundRequestBody struct {
	Reason string `json:"reason" validate:"required"`
}
