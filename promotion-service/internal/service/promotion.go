package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "promotion-service/api/promotion/v1"
	"promotion-service/internal/biz"
)

type PromotionService struct {
	pb.UnimplementedPromotionServiceServer

	uc *biz.PromotionUseCase
}

func NewPromotionService(uc *biz.PromotionUseCase) *PromotionService {
	return &PromotionService{
		uc: uc,
	}
}

func (s *PromotionService) CreateCoupon(ctx context.Context, req *pb.CreateCouponRequest) (*pb.CouponReply, error) {
	c, err := s.uc.CreateCoupon(ctx, &biz.Coupon{
		Name:      req.Name,
		Type:      req.Type,
		Threshold: req.Threshold,
		Discount:  req.Discount,
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.CouponReply{
		Id:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		Threshold: c.Threshold,
		Discount:  c.Discount,
		StartTime: timestamppb.New(c.StartTime),
		EndTime:   timestamppb.New(c.EndTime),
	}, nil
}

func (s *PromotionService) ListCoupons(ctx context.Context, req *pb.ListCouponsRequest) (*pb.ListCouponsReply, error) {
	cs, total, err := s.uc.ListCoupons(ctx, req.PageNum, req.PageSize)
	if err != nil {
		return nil, err
	}
	reply := &pb.ListCouponsReply{Total: total}
	for _, c := range cs {
		reply.Results = append(reply.Results, &pb.CouponReply{
			Id:        c.ID,
			Name:      c.Name,
			Type:      c.Type,
			Threshold: c.Threshold,
			Discount:  c.Discount,
			StartTime: timestamppb.New(c.StartTime),
			EndTime:   timestamppb.New(c.EndTime),
		})
	}
	return reply, nil
}

func (s *PromotionService) GrantCoupon(ctx context.Context, req *pb.GrantCouponRequest) (*emptypb.Empty, error) {
	err := s.uc.GrantCoupon(ctx, req.UserId, req.CouponId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *PromotionService) UseCoupon(ctx context.Context, req *pb.UseCouponRequest) (*emptypb.Empty, error) {
	err := s.uc.UseCoupon(ctx, req.UserId, req.CouponId, req.OrderId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *PromotionService) CalculateDiscount(ctx context.Context, req *pb.CalculateDiscountRequest) (*pb.CalculateDiscountReply, error) {
	discount, final, err := s.uc.CalculateDiscount(ctx, req.UserId, req.TotalAmount, req.CouponIds)
	if err != nil {
		return nil, err
	}
	return &pb.CalculateDiscountReply{
		DiscountAmount: discount,
		FinalAmount:    final,
	}, nil
}
