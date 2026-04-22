package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "product-service/api/product/v1"
	"product-service/internal/biz"
)

type ProductService struct {
	pb.UnimplementedProductServiceServer

	puc *biz.ProductUseCase
}

func NewProductService(puc *biz.ProductUseCase) *ProductService {
	return &ProductService{
		puc: puc,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductReply, error) {
	p, err := s.puc.CreateProduct(ctx, &biz.Product{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		ImageURL:    req.ImageUrl,
		CategoryID:  req.CategoryId,
		Price:       req.Price,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ProductReply{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Content:     p.Content,
		ImageUrl:    p.ImageURL,
		CategoryId:  p.CategoryID,
		Price:       p.Price,
		Status:      p.Status,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductReply, error) {
	p, err := s.puc.UpdateProduct(ctx, &biz.Product{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		ImageURL:    req.ImageUrl,
		CategoryID:  req.CategoryId,
		Price:       req.Price,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &pb.ProductReply{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Content:     p.Content,
		ImageUrl:    p.ImageURL,
		CategoryId:  p.CategoryID,
		Price:       p.Price,
		Status:      p.Status,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	err := s.puc.DeleteProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductReply, error) {
	p, err := s.puc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.ProductReply{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Content:     p.Content,
		ImageUrl:    p.ImageURL,
		CategoryId:  p.CategoryID,
		Price:       p.Price,
		Status:      p.Status,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsReply, error) {
	ps, total, err := s.puc.ListProducts(ctx, req.PageNum, req.PageSize, req.CategoryId, req.Keyword)
	if err != nil {
		return nil, err
	}
	reply := &pb.ListProductsReply{
		Total: total,
	}
	for _, p := range ps {
		reply.Results = append(reply.Results, &pb.ProductReply{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Content:     p.Content,
			ImageUrl:    p.ImageURL,
			CategoryId:  p.CategoryID,
			Price:       p.Price,
			Status:      p.Status,
			CreatedAt:   timestamppb.New(p.CreatedAt),
			UpdatedAt:   timestamppb.New(p.UpdatedAt),
		})
	}
	return reply, nil
}

func (s *ProductService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CategoryReply, error) {
	c, err := s.puc.CreateCategory(ctx, &biz.Category{
		Name:     req.Name,
		ParentID: req.ParentId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CategoryReply{
		Id:        c.ID,
		Name:      c.Name,
		ParentId:  c.ParentID,
		Level:     c.Level,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}, nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.CategoryReply, error) {
	c, err := s.puc.UpdateCategory(ctx, &biz.Category{
		ID:       req.Id,
		Name:     req.Name,
		ParentID: req.ParentId,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CategoryReply{
		Id:        c.ID,
		Name:      c.Name,
		ParentId:  c.ParentID,
		Level:     c.Level,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}, nil
}

func (s *ProductService) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	err := s.puc.DeleteCategory(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ProductService) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.CategoryReply, error) {
	c, err := s.puc.GetCategory(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.CategoryReply{
		Id:        c.ID,
		Name:      c.Name,
		ParentId:  c.ParentID,
		Level:     c.Level,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}, nil
}

func (s *ProductService) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesReply, error) {
	cs, err := s.puc.ListCategories(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}
	reply := &pb.ListCategoriesReply{}
	for _, c := range cs {
		reply.Results = append(reply.Results, &pb.CategoryReply{
			Id:        c.ID,
			Name:      c.Name,
			ParentId:  c.ParentID,
			Level:     c.Level,
			CreatedAt: timestamppb.New(c.CreatedAt),
			UpdatedAt: timestamppb.New(c.UpdatedAt),
		})
	}
	return reply, nil
}
