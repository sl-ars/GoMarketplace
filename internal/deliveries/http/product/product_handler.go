package product

import (
	"encoding/json"
	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/reqresp"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productService *services.ProductService
	offerService   *services.OfferService
}

func NewProductHandler(productService *services.ProductService, offerService *services.OfferService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		offerService:   offerService,
	}
}

// @Summary Create product
// @Tags products
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body reqresp.ProductCreateRequest true "Product data"
// @Success 201 {object} reqresp.StandardResponse
// @Router /api/admin/products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req reqresp.ProductCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	id, err := h.productService.CreateProduct(r.Context(), req.Name, req.Description)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to create product", err.Error())
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, "Product created successfully", reqresp.ProductCreateResponse{ID: id})
}

// @Summary Get product with offers
// @Tags products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/products/{id} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid product ID", err.Error())
		return
	}

	product, err := h.productService.GetProductByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch product", err.Error())
		return
	}

	offers, err := h.offerService.ListOffersByProduct(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch offers", err.Error())
		return
	}

	var offerResponses []reqresp.OfferShortResponse
	for _, o := range offers {
		offerResponses = append(offerResponses, reqresp.OfferShortResponse{
			ID:          o.ID,
			SellerID:    o.SellerID,
			Price:       o.Price,
			Stock:       o.Stock,
			IsAvailable: o.IsAvailable,
		})
	}

	response := reqresp.ProductWithOffersResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Offers:      offerResponses,
	}

	httpx.WriteSuccess(w, http.StatusOK, "Product fetched successfully", response)
}

// @Summary List all products
// @Tags products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page := 1
	pageSize := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	products, total, err := h.productService.ListProducts(r.Context(), page, pageSize)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch products", err.Error())
		return
	}

	var productResponses []reqresp.ProductResponse
	for _, p := range products {
		productResponses = append(productResponses, reqresp.ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response := reqresp.PaginatedResponse[reqresp.ProductResponse]{
		Items:      productResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	httpx.WriteSuccess(w, http.StatusOK, "Products fetched successfully", response)
}
