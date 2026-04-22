package biz

import (
	"context"
	"sync"
	"time"

	"user-service/internal/conf"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64
	Username  string
	Nickname  string
	Email     string
	Mobile    string
	Avatar    string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Address struct {
	ID        int64
	UserID    int64
	Name      string
	Mobile    string
	Province  string
	City      string
	District  string
	Detail    string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepo interface {
	CreateUser(ctx context.Context, u *User, password string) (*User, error)
	GetUser(ctx context.Context, id int64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, u *User) (*User, error)
}

type AddressRepo interface {
	CreateAddress(ctx context.Context, a *Address) (*Address, error)
	UpdateAddress(ctx context.Context, a *Address) (*Address, error)
	DeleteAddress(ctx context.Context, id int64) error
	GetAddress(ctx context.Context, id int64) (*Address, error)
	ListAddresses(ctx context.Context, userID int64) ([]*Address, error)
}

type UserUseCase struct {
	userRepo    UserRepo
	addressRepo AddressRepo
	tokenGen    *TokenGenerator

	mu         sync.Mutex
	loginFails map[string]int
}

func NewUserUseCase(userRepo UserRepo, addressRepo AddressRepo, tokenGen *TokenGenerator) *UserUseCase {
	return &UserUseCase{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		tokenGen:    tokenGen,
		loginFails:  make(map[string]int),
	}
}

func (uc *UserUseCase) Register(ctx context.Context, u *User, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return uc.userRepo.CreateUser(ctx, u, string(hashedPassword))
}

func (uc *UserUseCase) GetUser(ctx context.Context, id int64) (*User, error) {
	return uc.userRepo.GetUser(ctx, id)
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, u *User) (*User, error) {
	return uc.userRepo.UpdateUser(ctx, u)
}

func (uc *UserUseCase) Login(ctx context.Context, username, password string) (string, *User, error) {
	uc.mu.Lock()
	fails := uc.loginFails[username]
	uc.mu.Unlock()

	if fails > 5 {
		return "", nil, conf.ErrLoginLocked
	}

	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, conf.ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		uc.mu.Lock()
		uc.loginFails[username]++
		uc.mu.Unlock()
		return "", nil, conf.ErrInvalidPassword
	}

	uc.mu.Lock()
	uc.loginFails[username] = 0
	uc.mu.Unlock()

	token, err := uc.tokenGen.GenerateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (uc *UserUseCase) CreateAddress(ctx context.Context, a *Address) (*Address, error) {
	return uc.addressRepo.CreateAddress(ctx, a)
}

func (uc *UserUseCase) UpdateAddress(ctx context.Context, a *Address) (*Address, error) {
	return uc.addressRepo.UpdateAddress(ctx, a)
}

func (uc *UserUseCase) DeleteAddress(ctx context.Context, id int64) error {
	return uc.addressRepo.DeleteAddress(ctx, id)
}

func (uc *UserUseCase) GetAddress(ctx context.Context, id int64) (*Address, error) {
	return uc.addressRepo.GetAddress(ctx, id)
}

func (uc *UserUseCase) ListAddresses(ctx context.Context, userID int64) ([]*Address, error) {
	return uc.addressRepo.ListAddresses(ctx, userID)
}
