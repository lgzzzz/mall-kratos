package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type Payment struct {
	ID            string
	OrderID       string
	Amount        int64
	Currency      string
	Status        int32
	TransactionID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentRepo interface {
	Create(ctx context.Context, p *Payment) (*Payment, error)
	Get(ctx context.Context, id string) (*Payment, error)
	Update(ctx context.Context, p *Payment) (*Payment, error)
}

type PaymentUseCase struct {
	repo PaymentRepo
	log  *log.Helper
}

func NewPaymentUseCase(repo PaymentRepo, logger log.Logger) *PaymentUseCase {
	return &PaymentUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *PaymentUseCase) CreatePayment(ctx context.Context, p *Payment) (*Payment, error) {
	p.Status = 1 // Pending
	return uc.repo.Create(ctx, p)
}

func (uc *PaymentUseCase) GetPayment(ctx context.Context, id string) (*Payment, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *PaymentUseCase) Callback(ctx context.Context, id string, status int32, transactionID string) error {
	p, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	p.Status = status
	p.TransactionID = transactionID
	_, err = uc.repo.Update(ctx, p)
	return err
}

func (uc *PaymentUseCase) Refund(ctx context.Context, id string, amount int64) (*Payment, error) {
	p, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Status = 4 // Refunded
	return uc.repo.Update(ctx, p)
}
