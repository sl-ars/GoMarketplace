package usecases

import (
	"context"
	"go-app-marketplace/pkg/domain"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *domain.Product) (int64, error)
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	ListProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, error)
	GetTotalProducts(ctx context.Context) (int64, error)
}

type ProductUseCase struct {
	repo ProductRepository
}

func NewProductUseCase(repo ProductRepository) *ProductUseCase {
	return &ProductUseCase{repo: repo}
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, product *domain.Product) (int64, error) {
	return uc.repo.CreateProduct(ctx, product)
}

func (uc *ProductUseCase) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	return uc.repo.GetProductByID(ctx, id)
}

func (uc *ProductUseCase) ListProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, error) {
	return uc.repo.ListProducts(ctx, page, pageSize)
}

func (uc *ProductUseCase) GetTotalProducts(ctx context.Context) (int64, error) {
	return uc.repo.GetTotalProducts(ctx)
}
