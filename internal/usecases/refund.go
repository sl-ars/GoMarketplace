package usecases

import (
	"context"
	"go-app-marketplace/internal/repositories"
	"go-app-marketplace/pkg/domain"
)

type RefundUsecase struct {
	refundRepo *repositories.RefundRepository
	orderRepo  *repositories.OrderRepository
}

func NewRefundUsecase(r *repositories.RefundRepository, o *repositories.OrderRepository) *RefundUsecase {
	return &RefundUsecase{r, o}
}

func (u *RefundUsecase) RequestRefund(
	ctx context.Context, customerID, orderItemID int64, reason string) (int64, error) {

	item, err := u.orderRepo.GetOrderItemByID(ctx, orderItemID)
	if err != nil {
		return 0, err
	}
	if item.Status != domain.OrderItemStatusDelivered {
		return 0, repositories.ErrRefundStatusForbidden
	}
	if item.OrderUserID != customerID { // ensure owner
		return 0, repositories.ErrRefundStatusForbidden
	}
	return u.refundRepo.Create(ctx, *item, float64(item.Quantity)*item.UnitPrice, reason)
}

func (u *RefundUsecase) ApproveRefund(ctx context.Context, sellerID, refundID int64, approve bool) error {
	refund, err := u.refundRepo.GetByID(ctx, refundID)
	if err != nil {
		return err
	}
	if refund.SellerID != sellerID {
		return repositories.ErrRefundStatusForbidden
	}
	next := domain.RefundRejected
	if approve {
		next = domain.RefundApproved
	}
	return u.refundRepo.UpdateStatus(ctx, refundID, next)
}
