package mysql

import (
	"context"
	"errors"
	"fmt"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
	"gorm.io/gorm"
)

var (
	// ErrThirdPartyBindNotFound 第三方绑定不存在错误
	ErrThirdPartyBindNotFound = errors.New("third party bind not found")
	// ErrThirdPartyBindAlreadyExists 第三方绑定已存在错误
	ErrThirdPartyBindAlreadyExists = errors.New("third party bind already exists")
)

// ThirdPartyBindDAOImpl ThirdPartyBindDAO 的 MySQL 实现
type ThirdPartyBindDAOImpl struct {
	db *gorm.DB
}

// NewThirdPartyBindDAOImpl 创建 ThirdPartyBindDAOImpl 实例
func NewThirdPartyBindDAOImpl(db *gorm.DB) *ThirdPartyBindDAOImpl {
	return &ThirdPartyBindDAOImpl{
		db: db,
	}
}

// Insert 插入第三方绑定
func (dao *ThirdPartyBindDAOImpl) Insert(ctx context.Context, bind *entity.ThirdPartyBind) error {
	// 将领域对象转换为持久化对象
	po := dao.bindToPO(bind)

	// 使用 GORM 插入数据
	result := dao.db.WithContext(ctx).Create(po)
	if result.Error != nil {
		// 检查唯一约束冲突
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrThirdPartyBindAlreadyExists
		}
		return fmt.Errorf("failed to insert third party bind: %w", result.Error)
	}

	return nil
}

// Delete 删除第三方绑定
func (dao *ThirdPartyBindDAOImpl) Delete(ctx context.Context, userId int64, provider valueobject.OAuthProvider) error {
	result := dao.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userId, int8(provider)).
		Delete(&ThirdPartyBindPO{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete third party bind: %w", result.Error)
	}

	// 检查是否有行被删除
	if result.RowsAffected == 0 {
		return ErrThirdPartyBindNotFound
	}

	return nil
}

// FindByProvider 根据第三方提供商和提供商ID查找绑定
func (dao *ThirdPartyBindDAOImpl) FindByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (*entity.ThirdPartyBind, error) {
	var po ThirdPartyBindPO
	result := dao.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", int8(provider), providerId).
		First(&po)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrThirdPartyBindNotFound
		}
		return nil, fmt.Errorf("failed to find third party bind by provider: %w", result.Error)
	}

	// 将持久化对象转换为领域对象
	return dao.poToBind(&po), nil
}

// FindByUserId 根据用户ID查找所有第三方绑定
func (dao *ThirdPartyBindDAOImpl) FindByUserId(ctx context.Context, userId int64) ([]*entity.ThirdPartyBind, error) {
	var pos []ThirdPartyBindPO
	result := dao.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Find(&pos)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find third party binds by user id: %w", result.Error)
	}

	// 将持久化对象列表转换为领域对象列表
	binds := make([]*entity.ThirdPartyBind, 0, len(pos))
	for i := range pos {
		binds = append(binds, dao.poToBind(&pos[i]))
	}

	return binds, nil
}

// ExistsByProvider 检查第三方绑定是否已存在
func (dao *ThirdPartyBindDAOImpl) ExistsByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (bool, error) {
	var count int64
	result := dao.db.WithContext(ctx).
		Model(&ThirdPartyBindPO{}).
		Where("provider = ? AND provider_id = ?", int8(provider), providerId).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check third party bind existence: %w", result.Error)
	}
	return count > 0, nil
}

// bindToPO 将领域对象 ThirdPartyBind 转换为持久化对象 ThirdPartyBindPO
// 注意：这个方法应该从 converter 包调用，但为了避免循环依赖，直接在这里实现
func (dao *ThirdPartyBindDAOImpl) bindToPO(bind *entity.ThirdPartyBind) *ThirdPartyBindPO {
	po := &ThirdPartyBindPO{
		Id:         bind.Id(),
		UserId:     bind.UserId(),
		Provider:   int8(bind.Provider()),
		ProviderId: bind.ProviderId(),
		CreatedAt:  bind.CreatedAt(),
	}

	if bind.ProviderName() != "" {
		providerName := bind.ProviderName()
		po.ProviderName = &providerName
	}

	return po
}

// poToBind 将持久化对象 ThirdPartyBindPO 转换为领域对象 ThirdPartyBind
// 注意：这个方法应该从 converter 包调用，但为了避免循环依赖，直接在这里实现
func (dao *ThirdPartyBindDAOImpl) poToBind(po *ThirdPartyBindPO) *entity.ThirdPartyBind {
	bind := entity.NewThirdPartyBind(
		po.UserId,
		valueobject.OAuthProvider(po.Provider),
		po.ProviderId,
		"",
	)

	// 设置 ID
	bind.SetId(po.Id)

	// 设置 ProviderName
	if po.ProviderName != nil {
		bind.SetProviderName(*po.ProviderName)
	}

	// 设置 CreatedAt
	bind.SetCreatedAt(po.CreatedAt)

	return bind
}
