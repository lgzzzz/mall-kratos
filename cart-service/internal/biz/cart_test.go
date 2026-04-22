package biz

import (
	"context"
	"testing"

	"cart-service/internal/conf"
)

type mockProductRepo struct {
	products map[int64]*Product
}

func (m *mockProductRepo) GetProduct(ctx context.Context, id int64) (*Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, conf.ErrProductNotFound
	}
	return p, nil
}

type mockCartRepo struct {
	carts  map[int64]*Cart
	nextID int64
}

func (m *mockCartRepo) AddCart(ctx context.Context, c *Cart) (*Cart, error) {
	// Check if same user+product already exists
	for _, existing := range m.carts {
		if existing.UserID == c.UserID && existing.ProductID == c.ProductID {
			existing.Quantity += c.Quantity
			return existing, nil
		}
	}
	m.nextID++
	c.ID = m.nextID
	m.carts[c.ID] = c
	return c, nil
}

func (m *mockCartRepo) UpdateCart(ctx context.Context, c *Cart) (*Cart, error) {
	existing, ok := m.carts[c.ID]
	if !ok {
		return nil, conf.ErrCartNotFound
	}
	existing.Quantity = c.Quantity
	existing.Selected = c.Selected
	return existing, nil
}

func (m *mockCartRepo) DeleteCart(ctx context.Context, id int64) error {
	if _, ok := m.carts[id]; !ok {
		return conf.ErrCartNotFound
	}
	delete(m.carts, id)
	return nil
}

func (m *mockCartRepo) ListCart(ctx context.Context, userID int64) ([]*Cart, error) {
	var results []*Cart
	for _, c := range m.carts {
		if c.UserID == userID {
			results = append(results, c)
		}
	}
	return results, nil
}

func (m *mockCartRepo) ClearCart(ctx context.Context, userID int64) error {
	for id, c := range m.carts {
		if c.UserID == userID {
			delete(m.carts, id)
		}
	}
	return nil
}

func TestCartUseCase_AddCart_Success(t *testing.T) {
	productRepo := &mockProductRepo{
		products: map[int64]*Product{
			1: {ID: 1, Name: "Test Product", Price: 1000, Status: 1},
		},
	}
	cartRepo := &mockCartRepo{carts: make(map[int64]*Cart)}
	uc := NewCartUseCase(cartRepo, productRepo, nil)

	cart := &Cart{
		UserID:    1,
		ProductID: 1,
		Quantity:  2,
		Selected:  true,
	}

	result, err := uc.AddCart(context.Background(), cart)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", result.Quantity)
	}
}

func TestCartUseCase_AddCart_QuantityExceedsLimit(t *testing.T) {
	productRepo := &mockProductRepo{
		products: map[int64]*Product{
			1: {ID: 1, Name: "Test Product", Price: 1000, Status: 1},
		},
	}
	cartRepo := &mockCartRepo{carts: make(map[int64]*Cart)}
	uc := NewCartUseCase(cartRepo, productRepo, nil)

	cart := &Cart{
		UserID:    1,
		ProductID: 1,
		Quantity:  100, // exceeds maxCartQuantity (99)
		Selected:  true,
	}

	_, err := uc.AddCart(context.Background(), cart)
	if err != conf.ErrQuantityExceeded {
		t.Fatalf("expected ErrQuantityExceeded, got %v", err)
	}
}

func TestCartUseCase_AddCart_ProductNotFound(t *testing.T) {
	productRepo := &mockProductRepo{products: make(map[int64]*Product)}
	cartRepo := &mockCartRepo{carts: make(map[int64]*Cart)}
	uc := NewCartUseCase(cartRepo, productRepo, nil)

	cart := &Cart{
		UserID:    1,
		ProductID: 999,
		Quantity:  1,
		Selected:  true,
	}

	_, err := uc.AddCart(context.Background(), cart)
	if err != conf.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestCartUseCase_AddCart_ProductNotOnSale(t *testing.T) {
	productRepo := &mockProductRepo{
		products: map[int64]*Product{
			1: {ID: 1, Name: "Test Product", Price: 1000, Status: 0}, // not on sale
		},
	}
	cartRepo := &mockCartRepo{carts: make(map[int64]*Cart)}
	uc := NewCartUseCase(cartRepo, productRepo, nil)

	cart := &Cart{
		UserID:    1,
		ProductID: 1,
		Quantity:  1,
		Selected:  true,
	}

	_, err := uc.AddCart(context.Background(), cart)
	if err != conf.ErrProductNotFound {
		t.Fatalf("expected ErrProductNotFound for off-sale product, got %v", err)
	}
}

func TestCartUseCase_UpdateCart_QuantityExceedsLimit(t *testing.T) {
	productRepo := &mockProductRepo{
		products: map[int64]*Product{
			1: {ID: 1, Name: "Test Product", Price: 1000, Status: 1},
		},
	}
	cartRepo := &mockCartRepo{
		carts: map[int64]*Cart{
			1: {ID: 1, UserID: 1, ProductID: 1, Quantity: 2, Selected: true},
		},
	}
	uc := NewCartUseCase(cartRepo, productRepo, nil)

	_, err := uc.UpdateCart(context.Background(), &Cart{
		ID:       1,
		Quantity: 100, // exceeds max
		Selected: true,
	})
	if err != conf.ErrQuantityExceeded {
		t.Fatalf("expected ErrQuantityExceeded, got %v", err)
	}
}

func TestCartUseCase_ClearCart(t *testing.T) {
	productRepo := &mockProductRepo{
		products: map[int64]*Product{
			1: {ID: 1, Name: "Test Product", Price: 1000, Status: 1},
		},
	}
	cartRepo := &mockCartRepo{
		carts: map[int64]*Cart{
			1: {ID: 1, UserID: 1, ProductID: 1, Quantity: 2, Selected: true},
			2: {ID: 2, UserID: 1, ProductID: 2, Quantity: 1, Selected: true},
			3: {ID: 3, UserID: 2, ProductID: 1, Quantity: 3, Selected: true},
		},
	}
	uc := NewCartUseCase(cartRepo, productRepo, nil)

	err := uc.ClearCart(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// User 1's carts should be cleared
	items, _ := uc.ListCart(context.Background(), 1)
	if len(items) != 0 {
		t.Errorf("expected 0 items for user 1, got %d", len(items))
	}

	// User 2's cart should remain
	items2, _ := uc.ListCart(context.Background(), 2)
	if len(items2) != 1 {
		t.Errorf("expected 1 item for user 2, got %d", len(items2))
	}
}
