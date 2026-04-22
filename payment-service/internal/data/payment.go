package data

import (
	"context"

	"payment-service/internal/biz"
	"payment-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type paymentRepo struct {
	data *Data
	log  *log.Helper
}

func NewPaymentRepo(data *Data, logger log.Logger) biz.PaymentRepo {
	return &paymentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *paymentRepo) Create(ctx context.Context, p *biz.Payment) (*biz.Payment, error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	po := &model.Payment{
		ID:       p.ID,
		OrderID:  p.OrderID,
		Amount:   p.Amount,
		Currency: p.Currency,
		Status:   p.Status,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	p.CreatedAt = po.CreatedAt
	p.UpdatedAt = po.UpdatedAt
	return p, nil
}

func (r *paymentRepo) Get(ctx context.Context, id string) (*biz.Payment, error) {
	var po model.Payment
	if err := r.data.db.WithContext(ctx).First(&po, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &biz.Payment{
		ID:            po.ID,
		OrderID:       po.OrderID,
		Amount:        po.Amount,
		Currency:      po.Currency,
		Status:        po.Status,
		TransactionID: po.TransactionID,
		CreatedAt:     po.CreatedAt,
		UpdatedAt:     po.UpdatedAt,
	}, nil
}

func (r *paymentRepo) Update(ctx context.Context, p *biz.Payment) (*biz.Payment, error) {
	po := &model.Payment{
		ID:            p.ID,
		OrderID:       p.OrderID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		Status:        p.Status,
		TransactionID: p.TransactionID,
	}
	if err := r.data.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	p.UpdatedAt = po.UpdatedAt
	return p, nil
}
