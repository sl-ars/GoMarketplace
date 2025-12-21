package product

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
	"go-app-marketplace/pkg/reqresp"
)

type ProductHandler struct {
	productService *services.ProductService
	offerService   *services.OfferService
	logger         *logger.Logger
}

func NewProductHandler(productService *services.ProductService, offerService *services.OfferService, log *logger.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		offerService:   offerService,
		logger:         log,
	}
}

// Note: CreateProduct is now handled by admin handler at /api/admin/products

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
		h.logger.WithError(err).Warn("Invalid product ID format")
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	product, err := h.productService.GetProductByID(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch product")
		httpx.WriteError(w, http.StatusNotFound, apperror.ErrProductNotFound, apperror.ErrNotFound)
		return
	}

	offers, err := h.offerService.ListOffersByProduct(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch offers for product")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrOfferFetchFailed, apperror.ErrInternalServer)
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
		h.logger.WithError(err).Error("Failed to fetch products")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrProductFetchFailed, apperror.ErrInternalServer)
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

// @Summary Search products
// @Description Search for products using full-text search. Returns 404 if no products found.
// @Tags products
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} reqresp.StandardResponse
// @Failure 404 {object} reqresp.StandardResponse "No products found"
// @Router /api/products/search [get]
func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, "search query is required")
		return
	}

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

	products, total, err := h.productService.SearchProducts(r.Context(), query, page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to search products")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, "search failed")
		return
	}

	// Return product not found error if no products found
	if total == 0 {
		h.logger.WithField("query", query).Warn("No products found for search query")
		httpx.WriteError(w, http.StatusNotFound, apperror.ErrNotFound, apperror.ErrProductNotFound)
		return
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	response := map[string]interface{}{
		"items":       products,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}

	httpx.WriteSuccess(w, http.StatusOK, "Products found", response)
}
