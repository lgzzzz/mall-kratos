package service

import (
	"net/http"
	"strconv"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"

	cartV1 "cart-service/api/cart/v1"
	invV1 "inventory-service/api/inventory/v1"
	orderV1 "order-service/api/proto/order/v1"
	paymentV1 "payment-service/api/payment/v1"
	productV1 "product-service/api/product/v1"
	promoV1 "promotion-service/api/promotion/v1"
	userV1 "user-service/api/user/v1"

	"gateway/internal/client"
	gwMiddleware "gateway/internal/middleware"
)

type GatewayService struct {
	clients *client.Clients
	log     *log.Helper
}

func NewGatewayService(clients *client.Clients, logger log.Logger) *GatewayService {
	return &GatewayService{
		clients: clients,
		log:     log.NewHelper(logger),
	}
}

func (s *GatewayService) RegisterRoutes(srv *kratoshttp.Server) {
	s.registerUserRoutes(srv)
	s.registerProductRoutes(srv)
	s.registerOrderRoutes(srv)
	s.registerCartRoutes(srv)
	s.registerPaymentRoutes(srv)
	s.registerInventoryRoutes(srv)
	s.registerPromotionRoutes(srv)
}

func (s *GatewayService) registerUserRoutes(srv *kratoshttp.Server) {
	r := srv.Route("/api/user/v1")
	r.POST("/register", s.handleUserRegister)
	r.POST("/login", s.handleUserLogin)
	r.GET("/userinfo", s.handleUserGetUserInfo)
	r.POST("/address", s.handleUserAddAddress)
	r.PUT("/address", s.handleUserUpdateAddress)
	r.DELETE("/address/{id}", s.handleUserDeleteAddress)
	r.GET("/address/{id}", s.handleUserGetAddress)
	r.GET("/addresses", s.handleUserListAddresses)
}

func (s *GatewayService) registerProductRoutes(srv *kratoshttp.Server) {
	r := srv.Route("/api/product/v1")
	r.POST("/products", s.handleProductCreate)
	r.PUT("/products/{id}", s.handleProductUpdate)
	r.DELETE("/products/{id}", s.handleProductDelete)
	r.GET("/products/{id}", s.handleProductGet)
	r.GET("/products", s.handleProductList)
	r.POST("/categories", s.handleCategoryCreate)
	r.PUT("/categories/{id}", s.handleCategoryUpdate)
	r.DELETE("/categories/{id}", s.handleCategoryDelete)
	r.GET("/categories/{id}", s.handleCategoryGet)
	r.GET("/categories", s.handleCategoryList)
}

func (s *GatewayService) registerOrderRoutes(srv *kratoshttp.Server) {
	r := srv.Route("/api/order/v1")
	r.POST("/orders", s.handleOrderCreate)
	r.GET("/orders/{id}", s.handleOrderGet)
	r.PUT("/orders/{id}/status", s.handleOrderUpdateStatus)
	r.POST("/orders/{id}/cancel", s.handleOrderCancel)
	r.GET("/orders", s.handleOrderList)
}

func (s *GatewayService) registerCartRoutes(srv *kratoshttp.Server) {
	r := srv.Route("/api/cart/v1")
	r.POST("/cart", s.handleCartAdd)
	r.PUT("/cart/{id}", s.handleCartUpdate)
	r.DELETE("/cart/{id}", s.handleCartDelete)
	r.GET("/cart", s.handleCartGet)
	r.POST("/cart/clear", s.handleCartClear)
}

func (s *GatewayService) registerPaymentRoutes(srv *kratoshttp.Server) {
	r := srv.Route("/api/payment/v1")
	r.POST("/payment", s.handlePaymentCreate)
	r.GET("/payment/{id}", s.handlePaymentGet)
	r.POST("/payment/callback", s.handlePaymentCallback)
	r.POST("/payment/{id}/refund", s.handlePaymentRefund)
}

func (s *GatewayService) registerInventoryRoutes(srv *kratoshttp.Server) {
	r := srv.Route("/api/inventory/v1")
	r.GET("/inventory/{id}", s.handleInventoryGet)
	r.POST("/inventory/lock", s.handleInventoryLock)
	r.POST("/inventory/unlock", s.handleInventoryUnlock)
	r.POST("/inventory/confirm", s.handleInventoryConfirm)
}

func (s *GatewayService) registerPromotionRoutes(srv *kratoshttp.Server) {
	r := srv.Route("/api/promotion/v1")
	r.POST("/coupons", s.handleCouponCreate)
	r.GET("/coupons", s.handleCouponList)
	r.POST("/coupons/grant", s.handleCouponGrant)
	r.POST("/coupons/use", s.handleCouponUse)
	r.POST("/calculate-discount", s.handleCalculateDiscount)
}

// ========== User Handlers ==========

func (s *GatewayService) handleUserRegister(ctx kratoshttp.Context) error {
	var req userV1.RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.User.Register(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleUserLogin(ctx kratoshttp.Context) error {
	var req userV1.LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.User.Login(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleUserGetUserInfo(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	req := &userV1.GetUserInfoRequest{Id: authInfo.UserID}
	resp, err := s.clients.User.GetUserInfo(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleUserAddAddress(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	var req userV1.AddAddressRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.UserId = authInfo.UserID
	resp, err := s.clients.User.AddAddress(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleUserUpdateAddress(ctx kratoshttp.Context) error {
	var req userV1.UpdateAddressRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.User.UpdateAddress(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleUserDeleteAddress(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid address id")))
	}
	req := &userV1.DeleteAddressRequest{Id: id}
	_, err = s.clients.User.DeleteAddress(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, successResponse())
}

func (s *GatewayService) handleUserGetAddress(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid address id")))
	}
	req := &userV1.GetAddressRequest{Id: id}
	resp, err := s.clients.User.GetAddress(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleUserListAddresses(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	req := &userV1.ListAddressesRequest{UserId: authInfo.UserID}
	resp, err := s.clients.User.ListAddresses(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Product Handlers ==========

func (s *GatewayService) handleProductCreate(ctx kratoshttp.Context) error {
	var req productV1.CreateProductRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Product.CreateProduct(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleProductUpdate(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid product id")))
	}
	var req productV1.UpdateProductRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.Id = id
	resp, err := s.clients.Product.UpdateProduct(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleProductDelete(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid product id")))
	}
	req := &productV1.DeleteProductRequest{Id: id}
	_, err = s.clients.Product.DeleteProduct(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, successResponse())
}

func (s *GatewayService) handleProductGet(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid product id")))
	}
	req := &productV1.GetProductRequest{Id: id}
	resp, err := s.clients.Product.GetProduct(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleProductList(ctx kratoshttp.Context) error {
	var req productV1.ListProductsRequest
	if err := ctx.BindQuery(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Product.ListProducts(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Category Handlers ==========

func (s *GatewayService) handleCategoryCreate(ctx kratoshttp.Context) error {
	var req productV1.CreateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Product.CreateCategory(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCategoryUpdate(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid category id")))
	}
	var req productV1.UpdateCategoryRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.Id = id
	resp, err := s.clients.Product.UpdateCategory(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCategoryDelete(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid category id")))
	}
	req := &productV1.DeleteCategoryRequest{Id: id}
	_, err = s.clients.Product.DeleteCategory(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, successResponse())
}

func (s *GatewayService) handleCategoryGet(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid category id")))
	}
	req := &productV1.GetCategoryRequest{Id: id}
	resp, err := s.clients.Product.GetCategory(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCategoryList(ctx kratoshttp.Context) error {
	var req productV1.ListCategoriesRequest
	if err := ctx.BindQuery(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Product.ListCategories(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Order Handlers ==========

func (s *GatewayService) handleOrderCreate(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	var req orderV1.CreateOrderRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.UserId = authInfo.UserID
	resp, err := s.clients.Order.CreateOrder(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleOrderGet(ctx kratoshttp.Context) error {
	orderID := ctx.Vars().Get("id")
	req := &orderV1.GetOrderRequest{OrderId: orderID}
	resp, err := s.clients.Order.GetOrder(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleOrderUpdateStatus(ctx kratoshttp.Context) error {
	orderID := ctx.Vars().Get("id")
	var req orderV1.UpdateOrderStatusRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.OrderId = orderID
	resp, err := s.clients.Order.UpdateOrderStatus(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleOrderCancel(ctx kratoshttp.Context) error {
	orderID := ctx.Vars().Get("id")
	var req orderV1.CancelOrderRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.OrderId = orderID
	resp, err := s.clients.Order.CancelOrder(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleOrderList(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	var req orderV1.ListOrdersRequest
	if err := ctx.BindQuery(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.UserId = strconv.FormatInt(authInfo.UserID, 10)
	resp, err := s.clients.Order.ListOrders(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Cart Handlers ==========

func (s *GatewayService) handleCartAdd(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	var req cartV1.AddCartRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.UserId = authInfo.UserID
	resp, err := s.clients.Cart.AddCart(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCartUpdate(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid cart item id")))
	}
	var req cartV1.UpdateCartRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.Id = id
	resp, err := s.clients.Cart.UpdateCart(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCartDelete(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid cart item id")))
	}
	req := &cartV1.DeleteCartRequest{Id: id}
	resp, err := s.clients.Cart.DeleteCart(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCartGet(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	req := &cartV1.ListCartRequest{UserId: authInfo.UserID}
	resp, err := s.clients.Cart.ListCart(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCartClear(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	req := &cartV1.ClearCartRequest{UserId: authInfo.UserID}
	resp, err := s.clients.Cart.ClearCart(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Payment Handlers ==========

func (s *GatewayService) handlePaymentCreate(ctx kratoshttp.Context) error {
	var req paymentV1.CreatePaymentRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Payment.CreatePayment(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handlePaymentGet(ctx kratoshttp.Context) error {
	paymentID := ctx.Vars().Get("id")
	req := &paymentV1.GetPaymentRequest{PaymentId: paymentID}
	resp, err := s.clients.Payment.GetPayment(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handlePaymentCallback(ctx kratoshttp.Context) error {
	var req paymentV1.PaymentCallbackRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	_, err := s.clients.Payment.PaymentCallback(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, successResponse())
}

func (s *GatewayService) handlePaymentRefund(ctx kratoshttp.Context) error {
	paymentID := ctx.Vars().Get("id")
	var req paymentV1.RefundRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.PaymentId = paymentID
	resp, err := s.clients.Payment.Refund(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Inventory Handlers ==========

func (s *GatewayService) handleInventoryGet(ctx kratoshttp.Context) error {
	id, err := strconv.ParseInt(ctx.Vars().Get("id"), 10, 64)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(errors.BadRequest("BAD_REQUEST", "invalid inventory id")))
	}
	req := &invV1.GetInventoryRequest{Id: id}
	resp, err := s.clients.Inventory.GetInventory(ctx.Request().Context(), req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleInventoryLock(ctx kratoshttp.Context) error {
	var req invV1.LockStockRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Inventory.LockStock(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleInventoryUnlock(ctx kratoshttp.Context) error {
	var req invV1.UnlockStockRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Inventory.UnlockStock(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleInventoryConfirm(ctx kratoshttp.Context) error {
	var req invV1.ConfirmStockRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Inventory.ConfirmStock(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Promotion Handlers ==========

func (s *GatewayService) handleCouponList(ctx kratoshttp.Context) error {
	var req promoV1.ListCouponsRequest
	if err := ctx.BindQuery(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Promotion.ListCoupons(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCouponCreate(ctx kratoshttp.Context) error {
	var req promoV1.CreateCouponRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	resp, err := s.clients.Promotion.CreateCoupon(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (s *GatewayService) handleCouponGrant(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	var req promoV1.GrantCouponRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.UserId = authInfo.UserID
	_, err := s.clients.Promotion.GrantCoupon(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, successResponse())
}

func (s *GatewayService) handleCouponUse(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	var req promoV1.UseCouponRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.UserId = authInfo.UserID
	_, err := s.clients.Promotion.UseCoupon(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, successResponse())
}

func (s *GatewayService) handleCalculateDiscount(ctx kratoshttp.Context) error {
	authInfo, ok := gwMiddleware.GetAuthInfo(ctx.Request().Context())
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, errorResponse(errors.Unauthorized("UNAUTHORIZED", "unauthorized")))
	}
	var req promoV1.CalculateDiscountRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}
	req.UserId = authInfo.UserID
	resp, err := s.clients.Promotion.CalculateDiscount(ctx.Request().Context(), &req)
	if err != nil {
		return handleGRPCError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, resp)
}

// ========== Helper Functions ==========

func getPathParam(ctx kratoshttp.Context, key string) string {
	return ctx.Vars().Get(key)
}

func bindJSON(ctx kratoshttp.Context, v interface{}) error {
	return ctx.Bind(v)
}

func returnJSON(ctx kratoshttp.Context, code int, v interface{}) error {
	return ctx.JSON(code, v)
}

func handleGRPCError(ctx kratoshttp.Context, err error) error {
	if err == nil {
		return nil
	}
	e := errors.FromError(err)
	return ctx.JSON(gwMiddleware.GrpcToHTTPCode(e.Code), map[string]interface{}{
		"code":    e.Code,
		"reason":  e.Reason,
		"message": e.Message,
	})
}

func errorResponse(err error) map[string]interface{} {
	if err == nil {
		return nil
	}
	return map[string]interface{}{
		"code":    http.StatusBadRequest,
		"message": err.Error(),
	}
}

func successResponse() map[string]interface{} {
	return map[string]interface{}{
		"code":    0,
		"message": "success",
	}
}
