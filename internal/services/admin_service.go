package services

import (
	"context"
	"errors"
	"fmt"

	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/logger"
)

type AdminService struct {
	userRepo    *repositories.UserPostgresRepo
	productRepo *repositories.ProductRepository
	outboxRepo  *repositories.ElasticsearchOutboxRepository
	logger      *logger.Logger
}

func NewAdminService(
	userRepo *repositories.UserPostgresRepo,
	productRepo *repositories.ProductRepository,
	outboxRepo *repositories.ElasticsearchOutboxRepository,
	log *logger.Logger,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		logger:      log,
	}
}

// ============================================================================
// User Management
// ============================================================================

// ListUsers returns a paginated list of users with optional filters
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int, role string, isBanned *bool) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.userRepo.ListUsers(ctx, page, pageSize, role, isBanned)
}

// GetUserByID returns a user by ID
func (s *AdminService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// BanUser bans a user
func (s *AdminService) BanUser(ctx context.Context, userID, adminID int64, reason string) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Cannot ban admin users
	if user.Role == domain.UserRoleAdmin {
		return errors.New("cannot ban admin users")
	}

	// Cannot ban yourself
	if userID == adminID {
		return errors.New("cannot ban yourself")
	}

	// Cannot ban already banned user
	if user.IsBanned {
		return errors.New("user is already banned")
	}

	if reason == "" {
		reason = "No reason provided"
	}

	s.logger.WithField("user_id", userID).WithField("admin_id", adminID).Info("Banning user")
	return s.userRepo.BanUser(ctx, userID, adminID, reason)
}

// UnbanUser removes a ban from a user
func (s *AdminService) UnbanUser(ctx context.Context, userID int64) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.IsBanned {
		return errors.New("user is not banned")
	}

	s.logger.WithField("user_id", userID).Info("Unbanning user")
	return s.userRepo.UnbanUser(ctx, userID)
}

// UpdateUserRole updates a user's role
func (s *AdminService) UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error {
	if !domain.IsValidRole(role) {
		return errors.New("invalid role")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Prevent changing admin role
	if user.Role == domain.UserRoleAdmin {
		return errors.New("cannot change admin role")
	}

	s.logger.WithField("user_id", userID).WithField("new_role", role).Info("Updating user role")
	return s.userRepo.UpdateUserRole(ctx, userID, role)
}

// DeleteUser permanently deletes a user
func (s *AdminService) DeleteUser(ctx context.Context, userID, adminID int64) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Cannot delete admin users
	if user.Role == domain.UserRoleAdmin {
		return errors.New("cannot delete admin users")
	}

	// Cannot delete yourself
	if userID == adminID {
		return errors.New("cannot delete yourself")
	}

	s.logger.WithField("user_id", userID).WithField("admin_id", adminID).Warn("Deleting user")
	return s.userRepo.DeleteUser(ctx, userID)
}

// GetUserStats returns user statistics
func (s *AdminService) GetUserStats(ctx context.Context) (map[string]int64, error) {
	return s.userRepo.GetUserStats(ctx)
}

// ============================================================================
// Product Management
// ============================================================================

// CreateProduct creates a new product
func (s *AdminService) CreateProduct(ctx context.Context, name, description string) (int64, error) {
	product := &domain.Product{
		Name:        name,
		Description: description,
	}

	id, err := s.productRepo.CreateProduct(ctx, product)
	if err != nil {
		return 0, err
	}

	// Insert event into outbox for ES sync
	if s.outboxRepo != nil {
		eventPayload := map[string]interface{}{
			"name":        name,
			"description": description,
		}
		_ = s.outboxRepo.InsertEvent(ctx, nil, "product", id, "created", eventPayload)
	}

	s.logger.WithField("product_id", id).Info("Product created by admin")
	return id, nil
}

// GetProductByID returns a product by ID
func (s *AdminService) GetProductByID(ctx context.Context, productID int64) (*domain.Product, error) {
	return s.productRepo.GetProductByID(ctx, productID)
}

// ListProducts returns a paginated list of products
func (s *AdminService) ListProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, err := s.productRepo.ListProducts(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.productRepo.GetTotalProducts(ctx)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// UpdateProduct updates an existing product
func (s *AdminService) UpdateProduct(ctx context.Context, productID int64, name, description string) error {
	// Check if product exists
	_, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	product := &domain.Product{
		ID:          productID,
		Name:        name,
		Description: description,
	}

	if err := s.productRepo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	// Insert event into outbox for ES sync
	if s.outboxRepo != nil {
		eventPayload := map[string]interface{}{
			"name":        name,
			"description": description,
		}
		_ = s.outboxRepo.InsertEvent(ctx, nil, "product", productID, "updated", eventPayload)
	}

	s.logger.WithField("product_id", productID).Info("Product updated by admin")
	return nil
}

// DeleteProduct deletes a product
func (s *AdminService) DeleteProduct(ctx context.Context, productID int64) error {
	// Check if product exists
	_, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	if err := s.productRepo.DeleteProduct(ctx, productID); err != nil {
		return err
	}

	// Insert event into outbox for ES sync
	if s.outboxRepo != nil {
		_ = s.outboxRepo.InsertEvent(ctx, nil, "product", productID, "deleted", nil)
	}

	s.logger.WithField("product_id", productID).Warn("Product deleted by admin")
	return nil
}

// SearchProducts searches products by name
func (s *AdminService) SearchProducts(ctx context.Context, query string, page, pageSize int) ([]*domain.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.productRepo.SearchProductsByName(ctx, query, page, pageSize)
}

// ============================================================================
// Dashboard / Statistics
// ============================================================================

// GetDashboardStats returns overall statistics for the admin dashboard
func (s *AdminService) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// User stats
	userStats, err := s.userRepo.GetUserStats(ctx)
	if err != nil {
		return nil, err
	}
	stats["users"] = userStats

	// Product stats
	totalProducts, err := s.productRepo.GetTotalProducts(ctx)
	if err != nil {
		return nil, err
	}
	stats["products"] = map[string]int64{
		"total": totalProducts,
	}

	// Outbox stats (if available)
	if s.outboxRepo != nil {
		outboxStats, err := s.outboxRepo.GetEventStats(ctx)
		if err == nil {
			stats["elasticsearch_sync"] = outboxStats
		}
	}

	return stats, nil
}
