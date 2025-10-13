package usecases

import (
	"context"
	"errors"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/pkg/domain"
	"go-app-marketplace/pkg/reqresp"
)

type OrderUsecase struct {
	orderRepo *repositories.OrderRepository
	cartRepo  *repositories.CartRepository
	offerRepo *repositories.OfferRepository
}

func NewOrderUsecase(
	orderRepo *repositories.OrderRepository,
	cartRepo *repositories.CartRepository,
	offerRepo *repositories.OfferRepository,
) *OrderUsecase {
	return &OrderUsecase{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
		offerRepo: offerRepo,
	}
}

func (u *OrderUsecase) Checkout(ctx context.Context, userID int64) (int64, float64, error) {
	// Get cart items
	cartItems, err := u.cartRepo.GetItems(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	if len(cartItems) == 0 {
		return 0, 0, errors.New("cart is empty")
	}

	var totalAmount float64
	var orderItems []domain.OrderItem

	for _, item := range cartItems {
		offer, err := u.offerRepo.GetOfferByID(ctx, item.OfferID)
		if err != nil {
			return 0, 0, err
		}

		if offer.Stock < item.Quantity {
			return 0, 0, errors.New("insufficient stock for offer")
		}

		orderItems = append(orderItems, domain.OrderItem{
			OfferID:   item.OfferID,
			ProductID: offer.ProductID,
			SellerID:  offer.SellerID,
			Quantity:  item.Quantity,
			UnitPrice: offer.Price,
		})

		totalAmount += offer.Price * float64(item.Quantity)
	}

	// Create order
	orderID, err := u.orderRepo.CreateOrder(ctx, userID, totalAmount, orderItems)
	if err != nil {
		return 0, 0, err
	}

	// Clear cart
	if err := u.cartRepo.ClearCart(ctx, userID); err != nil {
		return 0, 0, err
	}

	return orderID, totalAmount, nil
}

func (u *OrderUsecase) CancelOrderItem(ctx context.Context, userID, itemID int64) error {
	return u.orderRepo.CancelOrderItem(ctx, userID, itemID)
}

func (u *OrderUsecase) ListOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	return u.orderRepo.ListOrders(ctx, userID)
}

func (u *OrderUsecase) ListOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	return u.orderRepo.ListOrderItems(ctx, orderID)
}

func (u *OrderUsecase) GetOrderByID(ctx context.Context, orderID int64) (*domain.Order, []domain.OrderItem, error) {
	return u.orderRepo.GetOrderByID(ctx, orderID)
}
func (u *OrderUsecase) UpdatePaymentStatusByOrderID(ctx context.Context, orderIDStr string, status domain.PaymentStatus) error {
	return u.orderRepo.UpdatePaymentStatusByOrderID(ctx, orderIDStr, status)
}

func (u *OrderUsecase) SellerUpdateOrderItemStatus(
	ctx context.Context,
	sellerID int64,
	itemID int64,
	newStatus domain.OrderItemStatus,
) error {

	item, err := u.orderRepo.GetOrderItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item.SellerID != sellerID {
		return errors.New("access denied: not your order item")
	}

	if !(item.Status == domain.OrderItemStatusPending && newStatus == domain.OrderItemStatusProcessing) &&
		!(item.Status == domain.OrderItemStatusProcessing && newStatus == domain.OrderItemStatusDelivered) {
		return errors.New("invalid status transition")
	}

	return u.orderRepo.UpdateOrderItemStatus(ctx, itemID, newStatus)
}

func (u *OrderUsecase) ListSellerOrderItems(
	ctx context.Context,
	sellerID int64,
) ([]reqresp.SellerOrderItem, error) {
	return u.orderRepo.ListOrderItemsBySeller(ctx, sellerID)
}
