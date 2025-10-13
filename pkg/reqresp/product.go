package reqresp

type ProductCreateRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type ProductCreateResponse struct {
	ID int64 `json:"id"`
}

type ProductResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProductWithOffersResponse struct {
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Offers      []OfferShortResponse `json:"offers"`
}

type PaginationRequest struct {
	Page     int `json:"page" query:"page" validate:"min=1"`
	PageSize int `json:"page_size" query:"page_size" validate:"min=1,max=100"`
}

type PaginatedResponse[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}
