package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-app-marketplace/pkg/domain"
)

const ProductIndex = "products"

type ProductDocument struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MinPrice    float64 `json:"min_price,omitempty"`
	MaxPrice    float64 `json:"max_price,omitempty"`
	AvgPrice    float64 `json:"avg_price,omitempty"`
	TotalOffers int     `json:"total_offers"`
}

type ProductSearchService struct {
	client *Client
}

func NewProductSearchService(client *Client) *ProductSearchService {
	return &ProductSearchService{client: client}
}

func (s *ProductSearchService) InitializeIndex(ctx context.Context) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "long",
				},
				"name": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"min_price": map[string]interface{}{
					"type": "float",
				},
				"max_price": map[string]interface{}{
					"type": "float",
				},
				"avg_price": map[string]interface{}{
					"type": "float",
				},
				"total_offers": map[string]interface{}{
					"type": "integer",
				},
			},
		},
	}

	return s.client.CreateIndex(ctx, ProductIndex, mapping)
}

func (s *ProductSearchService) IndexProduct(ctx context.Context, product *domain.Product, offers []*domain.Offer) error {
	doc := ProductDocument{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		TotalOffers: len(offers),
	}

	// Calculate price statistics from offers
	if len(offers) > 0 {
		var sum float64
		var count int
		minPrice := offers[0].Price
		maxPrice := offers[0].Price

		for _, offer := range offers {
			if offer.IsAvailable {
				sum += offer.Price
				count++
				if offer.Price < minPrice {
					minPrice = offer.Price
				}
				if offer.Price > maxPrice {
					maxPrice = offer.Price
				}
			}
		}

		if count > 0 {
			doc.MinPrice = minPrice
			doc.MaxPrice = maxPrice
			doc.AvgPrice = sum / float64(count)
		}
	}

	return s.client.Index(ctx, ProductIndex, strconv.FormatInt(product.ID, 10), doc)
}

func (s *ProductSearchService) SearchProducts(ctx context.Context, query string, page, pageSize int) ([]ProductDocument, int, error) {
	from := (page - 1) * pageSize

	searchQuery := map[string]interface{}{
		"from": from,
		"size": pageSize,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name^2", "description"},
				"type":   "best_fields",
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"_score": map[string]string{"order": "desc"},
			},
		},
	}

	data, err := s.client.Search(ctx, ProductIndex, searchQuery)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source ProductDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, 0, fmt.Errorf("error parsing search results: %w", err)
	}

	var products []ProductDocument
	for _, hit := range result.Hits.Hits {
		products = append(products, hit.Source)
	}

	return products, result.Hits.Total.Value, nil
}

func (s *ProductSearchService) DeleteProduct(ctx context.Context, productID int64) error {
	return s.client.Delete(ctx, ProductIndex, strconv.FormatInt(productID, 10))
}
