package repository

import (
	"context"
	"fmt"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
	domainerrors "github.com/yanking/price-watch/internal/auth/domain/errors"
	"github.com/yanking/price-watch/internal/auth/domain/repository"
	"github.com/yanking/price-watch/internal/auth/infrastructure/persistence/model"
	"gorm.io/gorm"
)

// UserRepositoryImpl UserRepository 的实现，直接操作 GORM
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepositoryImpl 创建 UserRepositoryImpl 实例
func NewUserRepositoryImpl(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

// Save 保存用户
func (r *UserRepositoryImpl) Save(ctx context.Context, user *entity.User) error {
	m := r.toModel(user)
	result := r.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		if isDuplicateKey(result.Error) {
			return domainerrors.ErrUserAlreadyExists
		}
		return fmt.Errorf("保存用户失败: %w", result.Error)
	}
	// 回写 ID
	user.SetID(m.Id)
	return nil
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	m := r.toModel(user)
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", m.Id).Updates(m)
	if result.Error != nil {
		return fmt.Errorf("更新用户失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrUserNotFound
	}
	return nil
}

// FindById 根据 ID 查找用户
func (r *UserRepositoryImpl) FindById(ctx context.Context, id int64) (*entity.User, error) {
	var m model.User
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", result.Error)
	}
	return r.toEntity(&m), nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var m model.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&m)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", result.Error)
	}
	return r.toEntity(&m), nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m model.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&m)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", result.Error)
	}
	return r.toEntity(&m), nil
}

// FindByPhone 根据区号和手机号查找用户
func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, areaCode, phone string) (*entity.User, error) {
	var m model.User
	result := r.db.WithContext(ctx).Where("area_code = ? AND phone = ?", areaCode, phone).First(&m)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %w", result.Error)
	}
	return r.toEntity(&m), nil
}

// ExistsByUsername 检查用户名是否已存在
func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("检查用户名失败: %w", result.Error)
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否已存在
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("检查邮箱失败: %w", result.Error)
	}
	return count > 0, nil
}

// ExistsByPhone 检查手机号是否已存在
func (r *UserRepositoryImpl) ExistsByPhone(ctx context.Context, areaCode, phone string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("area_code = ? AND phone = ?", areaCode, phone).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("检查手机号失败: %w", result.Error)
	}
	return count > 0, nil
}

// toModel 将领域实体转换为 GORM 模型
func (r *UserRepositoryImpl) toModel(user *entity.User) *model.User {
	m := &model.User{
		Id:            user.ID(),
		Username:      user.Username(),
		EmailVerified: user.EmailVerified(),
		PhoneVerified: user.PhoneVerified(),
		Status:        model.StatusToInt8(string(user.Status())),
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
	}

	if user.Password() != nil {
		hash := user.Password().Hash()
		m.PasswordHash = &hash
	}
	if user.Email() != nil {
		email := user.Email().Value()
		m.Email = &email
	}
	if user.AreaCode() != "" {
		areaCode := user.AreaCode()
		m.AreaCode = &areaCode
	}
	if user.Phone() != "" {
		phone := user.Phone()
		m.Phone = &phone
	}
	if user.Avatar() != "" {
		avatar := user.Avatar()
		m.Avatar = &avatar
	}
	if user.Nickname() != "" {
		nickname := user.Nickname()
		m.Nickname = &nickname
	}

	return m
}

// toEntity 将 GORM 模型转换为领域实体
func (r *UserRepositoryImpl) toEntity(m *model.User) *entity.User {
	var passwordHash string
	if m.PasswordHash != nil {
		passwordHash = *m.PasswordHash
	}
	var emailStr string
	if m.Email != nil {
		emailStr = *m.Email
	}
	var areaCodeStr string
	if m.AreaCode != nil {
		areaCodeStr = *m.AreaCode
	}
	var phoneStr string
	if m.Phone != nil {
		phoneStr = *m.Phone
	}
	var avatar string
	if m.Avatar != nil {
		avatar = *m.Avatar
	}
	var nickname string
	if m.Nickname != nil {
		nickname = *m.Nickname
	}

	user, err := entity.NewUserFromData(
		m.Id,
		m.Username,
		passwordHash,
		emailStr,
		areaCodeStr,
		phoneStr,
		m.EmailVerified,
		m.PhoneVerified,
		avatar,
		nickname,
		model.Int8ToStatus(m.Status),
		"", // oauthProvider
		"", // oauthID
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		// 数据来自数据库，理论上不会出错，降级处理
		user, _ = entity.NewUserFromData(
			m.Id, m.Username, "", "", "", "",
			m.EmailVerified, m.PhoneVerified, "", "",
			"active", "", "",
			m.CreatedAt, m.UpdatedAt,
		)
	}
	return user
}

// isDuplicateKey 检查是否为唯一约束冲突
func isDuplicateKey(err error) bool {
	return err == gorm.ErrDuplicatedKey
}

// 确保 UserRepositoryImpl 实现了 UserRepository 接口
var _ repository.UserRepository = (*UserRepositoryImpl)(nil)
