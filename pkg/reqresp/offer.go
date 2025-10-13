package reqresp

// OfferCreateRequest represents the payload to create a new offer
type OfferCreateRequest struct {
	// The ID of the product for which the offer is being created
	ProductID int64 `json:"product_id" validate:"required" example:"1" extensions:"x-order=1"`
	// The price of the offer
	Price float64 `json:"price" validate:"required" example:"29.99" extensions:"x-order=2"`
	// The available stock quantity
	Stock int `json:"stock" validate:"required" example:"100" extensions:"x-order=3"`
	// Whether the offer is currently available for purchase
	IsAvailable bool `json:"is_available" example:"true" extensions:"x-order=4"`
}

// OfferCreateResponse contains the ID of the newly created offer
type OfferCreateResponse struct {
	// The generated ID of the created offer
	ID int64 `json:"id" example:"1"`
}

// OfferUpdateRequest represents the payload to update an existing offer
type OfferUpdateRequest struct {
	// The updated price of the offer
	Price float64 `json:"price" validate:"required" example:"39.99" extensions:"x-order=1"`
	// The updated stock quantity
	Stock int `json:"stock" validate:"required" example:"50" extensions:"x-order=2"`
	// Whether the offer should be available for purchase
	IsAvailable bool `json:"is_available" example:"true" extensions:"x-order=3"`
}

// OfferResponse represents a complete offer object with all details
type OfferResponse struct {
	// The unique identifier of the offer
	ID int64 `json:"id" example:"1" extensions:"x-order=1"`
	// The product ID this offer is for
	ProductID int64 `json:"product_id" example:"42" extensions:"x-order=2"`
	// The ID of the seller who created the offer
	SellerID int64 `json:"seller_id" example:"5" extensions:"x-order=3"`
	// The price of the product in this offer
	Price float64 `json:"price" example:"29.99" extensions:"x-order=4"`
	// The available stock quantity
	Stock int `json:"stock" example:"100" extensions:"x-order=5"`
	// Whether the offer is currently available for purchase
	IsAvailable bool `json:"is_available" example:"true" extensions:"x-order=6"`
	// The timestamp when the offer was created, in RFC3339 format
	CreatedAt string `json:"created_at" example:"2023-04-15T14:32:20Z" extensions:"x-order=7"`
	// The timestamp when the offer was last updated, in RFC3339 format
	UpdatedAt string `json:"updated_at" example:"2023-04-16T09:12:55Z" extensions:"x-order=8"`
}

// OfferShortResponse represents a condensed view of an offer suitable for listing
type OfferShortResponse struct {
	// The unique identifier of the offer
	ID int64 `json:"id" example:"1" extensions:"x-order=1"`
	// The ID of the seller who created the offer
	SellerID int64 `json:"seller_id" example:"5" extensions:"x-order=2"`
	// The price of the product in this offer
	Price float64 `json:"price" example:"29.99" extensions:"x-order=3"`
	// The available stock quantity
	Stock int `json:"stock" example:"100" extensions:"x-order=4"`
	// Whether the offer is currently available for purchase
	IsAvailable bool `json:"is_available" example:"true" extensions:"x-order=5"`
}

// OfferFilterRequest represents filter parameters for listing offers
type OfferFilterRequest struct {
	// Minimum price to filter by
	MinPrice *float64 `json:"min_price,omitempty" query:"min_price" example:"10.00"`
	// Maximum price to filter by
	MaxPrice *float64 `json:"max_price,omitempty" query:"max_price" example:"100.00"`
	// Only show available offers
	OnlyAvailable *bool `json:"only_available,omitempty" query:"only_available" example:"true"`
	// Page number for pagination
	Page int `json:"page,omitempty" query:"page" validate:"min=1" example:"1" default:"1"`
	// Number of items per page
	PerPage int `json:"per_page,omitempty" query:"per_page" validate:"min=1,max=100" example:"20" default:"20"`
}

// OfferListResponse represents a paginated list of offers
type OfferListResponse struct {
	// List of offers
	Offers []OfferResponse `json:"offers" extensions:"x-order=1"`
	// Pagination metadata
	Pagination PaginationMetadata `json:"pagination" extensions:"x-order=2"`
}

// PaginationMetadata contains information about pagination
type PaginationMetadata struct {
	// Current page number
	CurrentPage int `json:"current_page" example:"1" extensions:"x-order=1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20" extensions:"x-order=2"`
	// Total number of items across all pages
	TotalItems int64 `json:"total_items" example:"42" extensions:"x-order=3"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"3" extensions:"x-order=4"`
	// Whether there is a next page available
	HasNextPage bool `json:"has_next_page" example:"true" extensions:"x-order=5"`
	// Whether there is a previous page available
	HasPrevPage bool `json:"has_prev_page" example:"false" extensions:"x-order=6"`
}
