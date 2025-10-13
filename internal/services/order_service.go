package services

import (
	"context"
	"errors"
	"go-app-marketplace/internal/usecases"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/reqresp"
)

type OrderService struct {
	orderUsecase   *usecases.OrderUsecase
	paymentService *PaymentService
}

func NewOrderService(orderUsecase *usecases.OrderUsecase) *OrderService {
	return &OrderService{
		orderUsecase: orderUsecase,
	}
}

// SetPaymentService sets the payment service - needed to avoid circular dependency
func (s *OrderService) SetPaymentService(paymentService *PaymentService) {
	s.paymentService = paymentService
}

func (s *OrderService) Checkout(ctx context.Context, userID int64) (*reqresp.CheckoutResponse, error) {
	orderID, totalAmount, err := s.orderUsecase.Checkout(ctx, userID)
	if err != nil {
		return nil, err
	}

	session, err := s.paymentService.CreateCheckoutSession(
		orderID,
		totalAmount,
		"https://localhost/payment-success",
		"https://localhost/payment-cancel",
	)
	if err != nil {
		return nil, err
	}

	return &reqresp.CheckoutResponse{
		OrderID:     orderID,
		TotalAmount: totalAmount,
		PaymentURL:  session.URL,
	}, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, userID, orderID int64) (*reqresp.OrderResponse, error) {
	order, items, err := s.orderUsecase.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Ensure the order belongs to the requesting user
	if order.UserID != userID {
		return nil, errors.New("order not found or access denied")
	}

	var itemResponses []reqresp.OrderItemResponse
	for _, item := range items {
		itemResponses = append(itemResponses, reqresp.OrderItemResponse{
			ID:        item.ID,
			OfferID:   item.OfferID,
			ProductID: item.ProductID,
			SellerID:  item.SellerID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Status:    string(item.Status),
		})
	}

	return &reqresp.OrderResponse{
		ID:            order.ID,
		UserID:        order.UserID,
		TotalAmount:   order.TotalAmount,
		Status:        string(order.Status),
		PaymentStatus: string(order.PaymentStatus),
		Items:         itemResponses,
	}, nil
}

func (s *OrderService) CheckoutExistingOrder(ctx context.Context, userID, orderID int64) (*reqresp.CheckoutResponse, error) {
	order, _, err := s.orderUsecase.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Ensure the order belongs to the requesting user
	if order.UserID != userID {
		return nil, errors.New("order not found or access denied")
	}

	// Check if payment is pending
	if order.PaymentStatus != domain.PaymentStatusPending {
		return nil, errors.New("order payment already processed")
	}

	// Create a new checkout session for the existing order
	session, err := s.paymentService.CreateCheckoutSession(
		orderID,
		order.TotalAmount,
		"https://localhost/payment-success",
		"https://localhost/payment-cancel",
	)

	if err != nil {
		return nil, err
	}

	return &reqresp.CheckoutResponse{
		OrderID:     orderID,
		TotalAmount: order.TotalAmount,
		PaymentURL:  session.URL,
	}, nil
}

func (s *OrderService) CancelOrderItem(ctx context.Context, userID, itemID int64) error {
	return s.orderUsecase.CancelOrderItem(ctx, userID, itemID)
}

func (s *OrderService) ListOrders(ctx context.Context, userID int64) ([]reqresp.OrderResponse, error) {
	orders, err := s.orderUsecase.ListOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	var resp []reqresp.OrderResponse
	for _, order := range orders {

		items, err := s.orderUsecase.ListOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		var itemResponses []reqresp.OrderItemResponse
		for _, item := range items {
			itemResponses = append(itemResponses, reqresp.OrderItemResponse{
				ID:        item.ID,
				OfferID:   item.OfferID,
				ProductID: item.ProductID,
				SellerID:  item.SellerID,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
				Status:    string(item.Status),
			})
		}

		resp = append(resp, reqresp.OrderResponse{
			ID:            order.ID,
			UserID:        order.UserID,
			TotalAmount:   order.TotalAmount,
			Status:        string(order.Status),
			PaymentStatus: string(order.PaymentStatus),
			Items:         itemResponses,
		})
	}

	return resp, nil
}

func (s *OrderService) UpdatePaymentStatusByOrderID(ctx context.Context, orderIDStr string, status domain.PaymentStatus) error {
	return s.orderUsecase.UpdatePaymentStatusByOrderID(ctx, orderIDStr, status)
}

func (s *OrderService) SellerUpdateOrderItemStatus(
	ctx context.Context,
	sellerID, itemID int64,
	status domain.OrderItemStatus,
) error {
	return s.orderUsecase.SellerUpdateOrderItemStatus(ctx, sellerID, itemID, status)
}

func (s *OrderService) ListSellerOrderItems(
	ctx context.Context,
	sellerID int64,
) ([]reqresp.SellerOrderItem, error) {
	return s.orderUsecase.ListSellerOrderItems(ctx, sellerID)
}
