package repositories

import (
	"context"
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
