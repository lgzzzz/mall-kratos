package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "user-service/api/user/v1"
	"user-service/internal/biz"
)

type UserService struct {
	pb.UnimplementedUserServiceServer

	uc *biz.UserUseCase
}

func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{
		uc: uc,
	}
}

func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserReply, error) {
	u, err := s.uc.Register(ctx, &biz.User{
		Username: req.Username,
		Mobile:   req.Mobile,
	}, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.UserReply{
		Id:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Mobile:    u.Mobile,
		Avatar:    u.Avatar,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	token, u, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginReply{
		Token: token,
		User: &pb.UserReply{
			Id:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Email:     u.Email,
			Mobile:    u.Mobile,
			Avatar:    u.Avatar,
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		},
	}, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.UserReply, error) {
	u, err := s.uc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.UserReply{
		Id:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Mobile:    u.Mobile,
		Avatar:    u.Avatar,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}, nil
}

func (s *UserService) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UserReply, error) {
	u, err := s.uc.UpdateUser(ctx, &biz.User{
		ID:       req.Id,
		Nickname: req.Nickname,
		Email:    req.Email,
		Avatar:   req.Avatar,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UserReply{
		Id:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Mobile:    u.Mobile,
		Avatar:    u.Avatar,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}, nil
}

func (s *UserService) AddAddress(ctx context.Context, req *pb.AddAddressRequest) (*pb.AddressReply, error) {
	a, err := s.uc.CreateAddress(ctx, &biz.Address{
		UserID:    req.UserId,
		Name:      req.Name,
		Mobile:    req.Mobile,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		return nil, err
	}
	return &pb.AddressReply{
		Id:        a.ID,
		UserId:    a.UserID,
		Name:      a.Name,
		Mobile:    a.Mobile,
		Province:  a.Province,
		City:      a.City,
		District:  a.District,
		Detail:    a.Detail,
		IsDefault: a.IsDefault,
		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}, nil
}

func (s *UserService) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.AddressReply, error) {
	a, err := s.uc.UpdateAddress(ctx, &biz.Address{
		ID:        req.Id,
		Name:      req.Name,
		Mobile:    req.Mobile,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	})
	if err != nil {
		return nil, err
	}
	return &pb.AddressReply{
		Id:        a.ID,
		UserId:    a.UserID,
		Name:      a.Name,
		Mobile:    a.Mobile,
		Province:  a.Province,
		City:      a.City,
		District:  a.District,
		Detail:    a.Detail,
		IsDefault: a.IsDefault,
		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}, nil
}

func (s *UserService) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*emptypb.Empty, error) {
	err := s.uc.DeleteAddress(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *UserService) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.AddressReply, error) {
	a, err := s.uc.GetAddress(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.AddressReply{
		Id:        a.ID,
		UserId:    a.UserID,
		Name:      a.Name,
		Mobile:    a.Mobile,
		Province:  a.Province,
		City:      a.City,
		District:  a.District,
		Detail:    a.Detail,
		IsDefault: a.IsDefault,
		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}, nil
}

func (s *UserService) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesReply, error) {
	as, err := s.uc.ListAddresses(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	reply := &pb.ListAddressesReply{}
	for _, a := range as {
		reply.Results = append(reply.Results, &pb.AddressReply{
			Id:        a.ID,
			UserId:    a.UserID,
			Name:      a.Name,
			Mobile:    a.Mobile,
			Province:  a.Province,
			City:      a.City,
			District:  a.District,
			Detail:    a.Detail,
			IsDefault: a.IsDefault,
			CreatedAt: timestamppb.New(a.CreatedAt),
			UpdatedAt: timestamppb.New(a.UpdatedAt),
		})
	}
	return reply, nil
}
