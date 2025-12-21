package repositories

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
	"go-app-marketplace/pkg/domain"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *domain.Product) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO products (name, description)
		VALUES ($1, $2)
		RETURNING id
	`, product.Name, product.Description).Scan(&id)
	return id, err
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	var product domain.Product
	err := r.db.GetContext(ctx, &product, `
		SELECT id, name, description, created_at, updated_at
		FROM products
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) ListProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, error) {
	var products []*domain.Product
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	return products, err
}

func (r *ProductRepository) GetTotalProducts(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM products
	`)
	return total, err
}

func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
	var products []*domain.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, created_at, updated_at
		FROM products
		ORDER BY id
	`)
	return products, err
}

// UpdateProduct updates an existing product
func (r *ProductRepository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE products 
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
	`, product.Name, product.Description, product.ID)
	return err
}

// DeleteProduct deletes a product by ID
func (r *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

// escapeLikePattern escapes special LIKE pattern characters to prevent LIKE injection
// This prevents attackers from using %, _, or \ to manipulate search patterns
func escapeLikePattern(s string) string {
	// Escape backslash first (since it's the escape character)
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Escape % (matches any sequence of characters)
	s = strings.ReplaceAll(s, `%`, `\%`)
	// Escape _ (matches any single character)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// SearchProductsByName searches products by name (case-insensitive)
// Security: Special LIKE characters (%, _, \) are escaped to prevent pattern injection
func (r *ProductRepository) SearchProductsByName(ctx context.Context, query string, page, pageSize int) ([]*domain.Product, int64, error) {
	offset := (page - 1) * pageSize

	// Escape special LIKE characters to prevent pattern injection (CWE-943)
	escapedQuery := escapeLikePattern(query)
	searchPattern := "%" + escapedQuery + "%"

	var products []*domain.Product
	err := r.db.SelectContext(ctx, &products, `
		SELECT id, name, description, created_at, updated_at
		FROM products
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.db.GetContext(ctx, &total, `
		SELECT COUNT(*)
		FROM products
		WHERE name ILIKE $1 OR description ILIKE $1
	`, searchPattern)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
