package data

import (
	"context"

	"user-service/internal/biz"
	"user-service/internal/model"

	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

type addressRepo struct {
	data *Data
	log  *log.Helper
}

func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func NewAddressRepo(data *Data, logger log.Logger) biz.AddressRepo {
	return &addressRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRepo) CreateUser(ctx context.Context, u *biz.User, password string) (*biz.User, error) {
	po := &model.User{
		Username: u.Username,
		Password: password, // 注意：实际应用中应加密存储
		Nickname: u.Nickname,
		Email:    u.Email,
		Mobile:   u.Mobile,
		Avatar:   u.Avatar,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	u.ID = po.ID
	u.CreatedAt = po.CreatedAt
	u.UpdatedAt = po.UpdatedAt
	return u, nil
}

func (r *userRepo) GetUser(ctx context.Context, id int64) (*biz.User, error) {
	var po model.User
	if err := r.data.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return &biz.User{
		ID:        po.ID,
		Username:  po.Username,
		Nickname:  po.Nickname,
		Email:     po.Email,
		Mobile:    po.Mobile,
		Avatar:    po.Avatar,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, u *biz.User) (*biz.User, error) {
	po := &model.User{
		ID:       u.ID,
		Nickname: u.Nickname,
		Email:    u.Email,
		Avatar:   u.Avatar,
	}
	if err := r.data.db.WithContext(ctx).Model(po).Updates(po).Error; err != nil {
		return nil, err
	}
	// 重新加载以获取更新后的完整数据
	return r.GetUser(ctx, u.ID)
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*biz.User, error) {
	var po model.User
	if err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		return nil, err
	}
	return &biz.User{
		ID:        po.ID,
		Username:  po.Username,
		Password:  po.Password,
		Nickname:  po.Nickname,
		Email:     po.Email,
		Mobile:    po.Mobile,
		Avatar:    po.Avatar,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}, nil
}

func (r *addressRepo) CreateAddress(ctx context.Context, a *biz.Address) (*biz.Address, error) {
	po := &model.Address{
		UserID:    a.UserID,
		Name:      a.Name,
		Mobile:    a.Mobile,
		Province:  a.Province,
		City:      a.City,
		District:  a.District,
		Detail:    a.Detail,
		IsDefault: a.IsDefault,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	a.ID = po.ID
	a.CreatedAt = po.CreatedAt
	a.UpdatedAt = po.UpdatedAt
	return a, nil
}

func (r *addressRepo) UpdateAddress(ctx context.Context, a *biz.Address) (*biz.Address, error) {
	po := &model.Address{
		ID:        a.ID,
		Name:      a.Name,
		Mobile:    a.Mobile,
		Province:  a.Province,
		City:      a.City,
		District:  a.District,
		Detail:    a.Detail,
		IsDefault: a.IsDefault,
	}
	if err := r.data.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	a.UpdatedAt = po.UpdatedAt
	return a, nil
}

func (r *addressRepo) DeleteAddress(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Delete(&model.Address{}, id).Error
}

func (r *addressRepo) GetAddress(ctx context.Context, id int64) (*biz.Address, error) {
	var po model.Address
	if err := r.data.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return &biz.Address{
		ID:        po.ID,
		UserID:    po.UserID,
		Name:      po.Name,
		Mobile:    po.Mobile,
		Province:  po.Province,
		City:      po.City,
		District:  po.District,
		Detail:    po.Detail,
		IsDefault: po.IsDefault,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}, nil
}

func (r *addressRepo) ListAddresses(ctx context.Context, userID int64) ([]*biz.Address, error) {
	var pos []model.Address
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error; err != nil {
		return nil, err
	}
	var res []*biz.Address
	for _, po := range pos {
		res = append(res, &biz.Address{
			ID:        po.ID,
			UserID:    po.UserID,
			Name:      po.Name,
			Mobile:    po.Mobile,
			Province:  po.Province,
			City:      po.City,
			District:  po.District,
			Detail:    po.Detail,
			IsDefault: po.IsDefault,
			CreatedAt: po.CreatedAt,
			UpdatedAt: po.UpdatedAt,
		})
	}
	return res, nil
}
