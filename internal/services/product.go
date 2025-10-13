package services

import (
	"context"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
)

type ProductService struct {
	usecase *usecases.ProductUseCase
}

func NewProductService(uc *usecases.ProductUseCase) *ProductService {
	return &ProductService{usecase: uc}
}

func (s *ProductService) CreateProduct(ctx context.Context, name, description string) (int64, error) {
	product := &domain.Product{
		Name:        name,
		Description: description,
	}
	return s.usecase.CreateProduct(ctx, product)
}

func (s *ProductService) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	return s.usecase.GetProductByID(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int64, error) {
	products, err := s.usecase.ListProducts(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.usecase.GetTotalProducts(ctx)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
