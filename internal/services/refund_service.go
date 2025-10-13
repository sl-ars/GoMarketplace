package services

import (
	"context"
	"go-app-marketplace/internal/usecases"
)

type RefundService struct{ uc *usecases.RefundUsecase }

func NewRefundService(uc *usecases.RefundUsecase) *RefundService { return &RefundService{uc} }

func (s *RefundService) Request(ctx context.Context, customerID, orderItemID int64, reason string) (int64, error) {
	return s.uc.RequestRefund(ctx, customerID, orderItemID, reason)
}

func (s *RefundService) Approve(ctx context.Context, sellerID, refundID int64, approve bool) error {
	return s.uc.ApproveRefund(ctx, sellerID, refundID, approve)
}
