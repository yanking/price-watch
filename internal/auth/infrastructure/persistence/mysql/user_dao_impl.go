package mysql

import (
	"context"
	"errors"
	"fmt"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"gorm.io/gorm"
)

var (
	// ErrUserNotFound 用户不存在错误
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists 用户已存在错误
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserDAOImpl UserDAO 的 MySQL 实现
type UserDAOImpl struct {
	db *gorm.DB
}

// NewUserDAOImpl 创建 UserDAOImpl 实例
func NewUserDAOImpl(db *gorm.DB) *UserDAOImpl {
	return &UserDAOImpl{
		db: db,
	}
}

// Insert 插入用户
func (dao *UserDAOImpl) Insert(ctx context.Context, user *entity.User) error {
	// 将领域对象转换为持久化对象
	po := dao.userToPO(user)

	// 使用 GORM 插入数据
	result := dao.db.WithContext(ctx).Create(po)
	if result.Error != nil {
		// 检查唯一约束冲突
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to insert user: %w", result.Error)
	}

	return nil
}

// Update 更新用户
func (dao *UserDAOImpl) Update(ctx context.Context, user *entity.User) error {
	// 将领域对象转换为持久化对象
	po := dao.userToPO(user)

	// 使用 GORM 更新数据
	result := dao.db.WithContext(ctx).Model(&UserPO{}).Where("id = ?", po.Id).Updates(po)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	// 检查是否有行被更新
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// FindById 根据 ID 查找用户
func (dao *UserDAOImpl) FindById(ctx context.Context, id int64) (*entity.User, error) {
	var po UserPO
	result := dao.db.WithContext(ctx).Where("id = ?", id).First(&po)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", result.Error)
	}

	// 将持久化对象转换为领域对象
	return dao.poToUser(&po), nil
}

// FindByUsername 根据用户名查找用户
func (dao *UserDAOImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var po UserPO
	result := dao.db.WithContext(ctx).Where("username = ?", username).First(&po)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by username: %w", result.Error)
	}

	// 将持久化对象转换为领域对象
	return dao.poToUser(&po), nil
}

// FindByEmail 根据邮箱查找用户
func (dao *UserDAOImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var po UserPO
	result := dao.db.WithContext(ctx).Where("email = ?", email).First(&po)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", result.Error)
	}

	// 将持久化对象转换为领域对象
	return dao.poToUser(&po), nil
}

// FindByPhone 根据区号和手机号查找用户
func (dao *UserDAOImpl) FindByPhone(ctx context.Context, areaCode, phone string) (*entity.User, error) {
	var po UserPO
	result := dao.db.WithContext(ctx).Where("area_code = ? AND phone = ?", areaCode, phone).First(&po)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by phone: %w", result.Error)
	}

	// 将持久化对象转换为领域对象
	return dao.poToUser(&po), nil
}

// ExistsByUsername 检查用户名是否已存在
func (dao *UserDAOImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	result := dao.db.WithContext(ctx).Model(&UserPO{}).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check username existence: %w", result.Error)
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否已存在
func (dao *UserDAOImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := dao.db.WithContext(ctx).Model(&UserPO{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check email existence: %w", result.Error)
	}
	return count > 0, nil
}

// ExistsByPhone 检查手机号是否已存在
func (dao *UserDAOImpl) ExistsByPhone(ctx context.Context, areaCode, phone string) (bool, error) {
	var count int64
	result := dao.db.WithContext(ctx).Model(&UserPO{}).Where("area_code = ? AND phone = ?", areaCode, phone).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check phone existence: %w", result.Error)
	}
	return count > 0, nil
}

// userToPO 将领域对象 User 转换为持久化对象 UserPO
// 注意：这个方法应该从 converter 包调用，但为了避免循环依赖，直接在这里实现
func (dao *UserDAOImpl) userToPO(user *entity.User) *UserPO {
	po := &UserPO{
		Id:            int64(user.ID()),
		Username:      user.Username(),
		EmailVerified: user.EmailVerified(),
		PhoneVerified: user.PhoneVerified(),
		Status:        userStatusToInt8(user.Status()),
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
	}

	// 处理密码哈希
	if user.Password() != nil {
		hash := user.Password().Hash()
		po.PasswordHash = &hash
	}

	// 处理邮箱
	if user.Email() != nil {
		email := user.Email().Value()
		po.Email = &email
	}

	// 处理区号
	if user.AreaCode() != "" {
		areaCode := user.AreaCode()
		po.AreaCode = &areaCode
	}

	// 处理手机号
	if user.Phone() != "" {
		phone := user.Phone()
		po.Phone = &phone
	}

	return po
}

// poToUser 将持久化对象 UserPO 转换为领域对象 User
// 注意：这个方法应该从 converter 包调用，但为了避免循环依赖，直接在这里实现
func (dao *UserDAOImpl) poToUser(po *UserPO) *entity.User {
	// 构建密码哈希字符串
	var passwordHash string
	if po.PasswordHash != nil {
		passwordHash = *po.PasswordHash
	}

	// 构建邮箱字符串
	var emailStr string
	if po.Email != nil {
		emailStr = *po.Email
	}

	// 构建区号字符串
	var areaCodeStr string
	if po.AreaCode != nil {
		areaCodeStr = *po.AreaCode
	}

	// 构建手机号字符串
	var phoneStr string
	if po.Phone != nil {
		phoneStr = *po.Phone
	}

	user, err := entity.NewUserFromData(
		uint64(po.Id),
		po.Username,
		passwordHash,
		emailStr,
		areaCodeStr,
		phoneStr,
		po.EmailVerified,
		po.PhoneVerified,
		string(int8ToUserStatus(po.Status)),
		"", // oauthProvider - 暂不处理
		"", // oauthID - 暂不处理
		po.CreatedAt,
		po.UpdatedAt,
	)
	if err != nil {
		// 理论上不应该发生，因为数据来自数据库
		// 如果发生，返回一个基本的用户对象
		user, _ = entity.NewUserFromData(
			uint64(po.Id),
			po.Username,
			"",
			"",
			"",
			"",
			po.EmailVerified,
			po.PhoneVerified,
			"active",
			"",
			"",
			po.CreatedAt,
			po.UpdatedAt,
		)
	}

	return user
}

// userStatusToInt8 将 UserStatus 转换为 int8
func userStatusToInt8(status entity.UserStatus) int8 {
	switch status {
	case entity.UserStatusActive:
		return 1
	case entity.UserStatusInactive:
		return 2
	case entity.UserStatusSuspended:
		return 3
	case entity.UserStatusDeleted:
		return 4
	default:
		return 1 // 默认为激活状态
	}
}

// int8ToUserStatus 将 int8 转换为 UserStatus
func int8ToUserStatus(status int8) entity.UserStatus {
	switch status {
	case 1:
		return entity.UserStatusActive
	case 2:
		return entity.UserStatusInactive
	case 3:
		return entity.UserStatusSuspended
	case 4:
		return entity.UserStatusDeleted
	default:
		return entity.UserStatusActive
	}
}
