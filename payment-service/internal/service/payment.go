package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "payment-service/api/payment/v1"
	"payment-service/internal/biz"
)

type PaymentService struct {
	pb.UnimplementedPaymentServiceServer

	uc *biz.PaymentUseCase
}

func NewPaymentService(uc *biz.PaymentUseCase) *PaymentService {
	return &PaymentService{
		uc: uc,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.PaymentReply, error) {
	p, err := s.uc.CreatePayment(ctx, &biz.Payment{
		OrderID:  req.OrderId,
		Amount:   req.Amount,
		Currency: req.Currency,
	})
	if err != nil {
		return nil, err
	}
	return &pb.PaymentReply{
		Id:        p.ID,
		OrderId:   p.OrderID,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    pb.PaymentStatus(p.Status),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.PaymentReply, error) {
	p, err := s.uc.GetPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, err
	}
	return &pb.PaymentReply{
		Id:        p.ID,
		OrderId:   p.OrderID,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    pb.PaymentStatus(p.Status),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}, nil
}

func (s *PaymentService) PaymentCallback(ctx context.Context, req *pb.PaymentCallbackRequest) (*emptypb.Empty, error) {
	err := s.uc.Callback(ctx, req.PaymentId, int32(req.Status), req.TransactionId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *PaymentService) Refund(ctx context.Context, req *pb.RefundRequest) (*pb.PaymentReply, error) {
	p, err := s.uc.Refund(ctx, req.PaymentId, req.Amount)
	if err != nil {
		return nil, err
	}
	return &pb.PaymentReply{
		Id:        p.ID,
		OrderId:   p.OrderID,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    pb.PaymentStatus(p.Status),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}, nil
}
