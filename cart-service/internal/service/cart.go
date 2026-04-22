package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "cart-service/api/cart/v1"
	"cart-service/internal/biz"
)

type CartService struct {
	pb.UnimplementedCartServiceServer

	uc *biz.CartUseCase
}

func NewCartService(uc *biz.CartUseCase) *CartService {
	return &CartService{
		uc: uc,
	}
}

func (s *CartService) AddCart(ctx context.Context, req *pb.AddCartRequest) (*pb.CartReply, error) {
	c, err := s.uc.AddCart(ctx, &biz.Cart{
		UserID:    req.UserId,
		ProductID: req.ProductId,
		Quantity:  req.Quantity,
		Selected:  true,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CartReply{
		Id:        c.ID,
		UserId:    c.UserID,
		ProductId: c.ProductID,
		Quantity:  c.Quantity,
		Selected:  c.Selected,
	}, nil
}

func (s *CartService) UpdateCart(ctx context.Context, req *pb.UpdateCartRequest) (*pb.CartReply, error) {
	c, err := s.uc.UpdateCart(ctx, &biz.Cart{
		ID:       req.Id,
		Quantity: req.Quantity,
		Selected: req.Selected,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CartReply{
		Id:        c.ID,
		UserId:    c.UserID,
		ProductId: c.ProductID,
		Quantity:  c.Quantity,
		Selected:  c.Selected,
	}, nil
}

func (s *CartService) DeleteCart(ctx context.Context, req *pb.DeleteCartRequest) (*emptypb.Empty, error) {
	err := s.uc.DeleteCart(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *CartService) ListCart(ctx context.Context, req *pb.ListCartRequest) (*pb.ListCartReply, error) {
	cs, err := s.uc.ListCart(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	reply := &pb.ListCartReply{}
	for _, c := range cs {
		reply.Results = append(reply.Results, &pb.CartReply{
			Id:        c.ID,
			UserId:    c.UserID,
			ProductId: c.ProductID,
			Quantity:  c.Quantity,
			Selected:  c.Selected,
		})
	}
	return reply, nil
}

func (s *CartService) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*emptypb.Empty, error) {
	err := s.uc.ClearCart(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
