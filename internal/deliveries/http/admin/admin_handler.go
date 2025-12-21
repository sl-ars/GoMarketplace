package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-app-marketplace/internal/services"
	"go-app-marketplace/pkg/apperror"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/httpx"
	"go-app-marketplace/pkg/logger"
)

type Handler struct {
	adminService *services.AdminService
	logger       *logger.Logger
}

func NewHandler(adminService *services.AdminService, log *logger.Logger) *Handler {
	return &Handler{
		adminService: adminService,
		logger:       log,
	}
}

// ============================================================================
// User Management Endpoints
// ============================================================================

// ListUsersRequest represents the query parameters for listing users
type ListUsersRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Role     string `json:"role"`
	IsBanned string `json:"is_banned"` // "true", "false", or empty
}

// @Summary List all users
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param role query string false "Filter by role (admin, seller, customer)"
// @Param is_banned query string false "Filter by ban status (true, false)"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/users [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	role := r.URL.Query().Get("role")
	isBannedStr := r.URL.Query().Get("is_banned")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var isBanned *bool
	if isBannedStr == "true" {
		b := true
		isBanned = &b
	} else if isBannedStr == "false" {
		b := false
		isBanned = &b
	}

	users, total, err := h.adminService.ListUsers(r.Context(), page, pageSize, role, isBanned)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrInternalServer)
		return
	}

	// Transform users to response (without sensitive fields)
	usersResp := make([]UserResponse, len(users))
	for i, u := range users {
		usersResp[i] = UserResponse{
			ID:            u.ID,
			Username:      u.Username,
			Email:         u.Email,
			Role:          string(u.Role),
			EmailVerified: u.EmailVerified,
			CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			IsBanned:      u.IsBanned,
			BanReason:     u.BanReason.String,
		}
		if u.BannedAt.Valid {
			usersResp[i].BannedAt = u.BannedAt.Time.Format("2006-01-02T15:04:05Z")
		}
	}

	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	httpx.WriteSuccess(w, http.StatusOK, "Users retrieved", map[string]interface{}{
		"users":       usersResp,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UserResponse represents user data returned to admin
type UserResponse struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
	IsBanned      bool   `json:"is_banned"`
	BannedAt      string `json:"banned_at,omitempty"`
	BanReason     string `json:"ban_reason,omitempty"`
}

// @Summary Get user by ID
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/users/{id} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	user, err := h.adminService.GetUserByID(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		httpx.WriteError(w, http.StatusNotFound, apperror.ErrNotFound, apperror.ErrUserNotFound)
		return
	}

	resp := UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Role:          string(user.Role),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		IsBanned:      user.IsBanned,
		BanReason:     user.BanReason.String,
	}
	if user.BannedAt.Valid {
		resp.BannedAt = user.BannedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	httpx.WriteSuccess(w, http.StatusOK, "User retrieved", resp)
}

// BanUserRequest represents the request to ban a user
type BanUserRequest struct {
	Reason string `json:"reason"`
}

// @Summary Ban a user
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param input body BanUserRequest true "Ban reason"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/users/{id}/ban [post]
func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	var req BanUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body (no reason provided)
		req.Reason = ""
	}

	// Get admin ID from context
	adminID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrUnauthorized)
		return
	}

	if err := h.adminService.BanUser(r.Context(), userID, adminID, req.Reason); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to ban user")

		switch err.Error() {
		case "cannot ban admin users":
			httpx.WriteError(w, http.StatusForbidden, apperror.ErrForbidden, apperror.ErrCannotBanAdmin)
		case "cannot ban yourself":
			httpx.WriteError(w, http.StatusForbidden, apperror.ErrForbidden, apperror.ErrCannotBanSelf)
		case "user is already banned":
			httpx.WriteError(w, http.StatusConflict, apperror.ErrConflict, apperror.ErrUserAlreadyBanned)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrBanFailed)
		}
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "User banned successfully", nil)
}

// @Summary Unban a user
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/users/{id}/unban [post]
func (h *Handler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	if err := h.adminService.UnbanUser(r.Context(), userID); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to unban user")

		if err.Error() == "user is not banned" {
			httpx.WriteError(w, http.StatusConflict, apperror.ErrConflict, apperror.ErrUserNotBanned)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrUnbanFailed)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "User unbanned successfully", nil)
}

// UpdateRoleRequest represents the request to update a user's role
type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// @Summary Update user role
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param input body UpdateRoleRequest true "New role"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/users/{id}/role [put]
func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidJSON)
		return
	}

	role := domain.UserRole(req.Role)
	if !domain.IsValidRole(role) {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidRole)
		return
	}

	if err := h.adminService.UpdateUserRole(r.Context(), userID, role); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user role")

		if err.Error() == "cannot change admin role" {
			httpx.WriteError(w, http.StatusForbidden, apperror.ErrForbidden, apperror.ErrCannotChangeAdminRole)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrRoleUpdateFailed)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "User role updated successfully", nil)
}

// @Summary Delete a user
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/users/{id} [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	// Get admin ID from context
	adminID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, apperror.ErrUnauthorized, apperror.ErrUnauthorized)
		return
	}

	if err := h.adminService.DeleteUser(r.Context(), userID, adminID); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to delete user")

		switch err.Error() {
		case "cannot delete admin users":
			httpx.WriteError(w, http.StatusForbidden, apperror.ErrForbidden, apperror.ErrCannotDeleteAdmin)
		case "cannot delete yourself":
			httpx.WriteError(w, http.StatusForbidden, apperror.ErrForbidden, apperror.ErrCannotDeleteSelf)
		default:
			httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrUserDeleteFailed)
		}
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "User deleted successfully", nil)
}

// ============================================================================
// Product Management Endpoints
// ============================================================================

// ProductRequest represents create/update product request
type ProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// @Summary Create a product
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body ProductRequest true "Product data"
// @Success 201 {object} reqresp.StandardResponse
// @Router /api/admin/products [post]
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidJSON)
		return
	}

	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, "Product name is required")
		return
	}

	id, err := h.adminService.CreateProduct(r.Context(), req.Name, req.Description)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create product")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrProductCreateFailed)
		return
	}

	httpx.WriteSuccess(w, http.StatusCreated, "Product created successfully", map[string]int64{"id": id})
}

// @Summary List all products
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/products [get]
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, total, err := h.adminService.ListProducts(r.Context(), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list products")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrProductFetchFailed)
		return
	}

	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	httpx.WriteSuccess(w, http.StatusOK, "Products retrieved", map[string]interface{}{
		"products":    products,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// @Summary Get product by ID
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/products/{id} [get]
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	product, err := h.adminService.GetProductByID(r.Context(), productID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get product")
		httpx.WriteError(w, http.StatusNotFound, apperror.ErrNotFound, apperror.ErrProductNotFound)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Product retrieved", product)
}

// @Summary Update a product
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param input body ProductRequest true "Product data"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/products/{id} [put]
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidJSON)
		return
	}

	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, "Product name is required")
		return
	}

	if err := h.adminService.UpdateProduct(r.Context(), productID, req.Name, req.Description); err != nil {
		h.logger.WithError(err).WithField("product_id", productID).Error("Failed to update product")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrProductUpdateFailed)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Product updated successfully", nil)
}

// @Summary Delete a product
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/products/{id} [delete]
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, apperror.ErrBadRequest, apperror.ErrInvalidID)
		return
	}

	if err := h.adminService.DeleteProduct(r.Context(), productID); err != nil {
		h.logger.WithError(err).WithField("product_id", productID).Error("Failed to delete product")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrProductDeleteFailed)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Product deleted successfully", nil)
}

// ============================================================================
// Dashboard / Statistics Endpoints
// ============================================================================

// @Summary Get admin dashboard statistics
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/dashboard [get]
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetDashboardStats(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get dashboard stats")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrInternalServer)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "Dashboard statistics", stats)
}

// @Summary Get user statistics
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} reqresp.StandardResponse
// @Router /api/admin/stats/users [get]
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.adminService.GetUserStats(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user stats")
		httpx.WriteError(w, http.StatusInternalServerError, apperror.ErrInternalServer, apperror.ErrInternalServer)
		return
	}

	httpx.WriteSuccess(w, http.StatusOK, "User statistics", stats)
}
