package repository

import (
	"context"
	"fmt"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
	domainerrors "github.com/yanking/price-watch/internal/auth/domain/errors"
	"github.com/yanking/price-watch/internal/auth/domain/repository"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
	"github.com/yanking/price-watch/internal/auth/infrastructure/persistence/model"
	"gorm.io/gorm"
)

// ThirdPartyBindRepositoryImpl ThirdPartyBindRepository 的实现
type ThirdPartyBindRepositoryImpl struct {
	db *gorm.DB
}

// NewThirdPartyBindRepositoryImpl 创建 ThirdPartyBindRepositoryImpl 实例
func NewThirdPartyBindRepositoryImpl(db *gorm.DB) *ThirdPartyBindRepositoryImpl {
	return &ThirdPartyBindRepositoryImpl{db: db}
}

// Save 保存第三方绑定
func (r *ThirdPartyBindRepositoryImpl) Save(ctx context.Context, bind *entity.ThirdPartyBind) error {
	m := r.toModel(bind)
	result := r.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		if isDuplicateKey(result.Error) {
			return domainerrors.ErrBindExists
		}
		return fmt.Errorf("保存绑定失败: %w", result.Error)
	}
	bind.SetId(m.Id)
	return nil
}

// Delete 删除第三方绑定
func (r *ThirdPartyBindRepositoryImpl) Delete(ctx context.Context, userId int64, provider valueobject.OAuthProvider) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userId, int8(provider)).
		Delete(&model.ThirdPartyBind{})
	if result.Error != nil {
		return fmt.Errorf("删除绑定失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrBindNotFound
	}
	return nil
}

// FindByProvider 根据第三方提供商和提供商ID查找绑定
func (r *ThirdPartyBindRepositoryImpl) FindByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (*entity.ThirdPartyBind, error) {
	var m model.ThirdPartyBind
	result := r.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", int8(provider), providerId).
		First(&m)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询绑定失败: %w", result.Error)
	}
	return r.toEntity(&m), nil
}

// FindByUserId 根据用户ID查找所有第三方绑定
func (r *ThirdPartyBindRepositoryImpl) FindByUserId(ctx context.Context, userId int64) ([]*entity.ThirdPartyBind, error) {
	var models []model.ThirdPartyBind
	result := r.db.WithContext(ctx).Where("user_id = ?", userId).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("查询绑定列表失败: %w", result.Error)
	}

	binds := make([]*entity.ThirdPartyBind, 0, len(models))
	for i := range models {
		binds = append(binds, r.toEntity(&models[i]))
	}
	return binds, nil
}

// ExistsByProvider 检查第三方绑定是否已存在
func (r *ThirdPartyBindRepositoryImpl) ExistsByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&model.ThirdPartyBind{}).
		Where("provider = ? AND provider_id = ?", int8(provider), providerId).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("检查绑定失败: %w", result.Error)
	}
	return count > 0, nil
}

// toModel 将领域实体转换为 GORM 模型
func (r *ThirdPartyBindRepositoryImpl) toModel(bind *entity.ThirdPartyBind) *model.ThirdPartyBind {
	m := &model.ThirdPartyBind{
		Id:         bind.Id(),
		UserId:     uint64(bind.UserId()),
		Provider:   int8(bind.Provider()),
		ProviderId: bind.ProviderId(),
		CreatedAt:  bind.CreatedAt(),
	}
	if bind.ProviderName() != "" {
		name := bind.ProviderName()
		m.ProviderName = &name
	}
	return m
}

// toEntity 将 GORM 模型转换为领域实体
func (r *ThirdPartyBindRepositoryImpl) toEntity(m *model.ThirdPartyBind) *entity.ThirdPartyBind {
	bind := entity.NewThirdPartyBind(
		int64(m.UserId),
		valueobject.OAuthProvider(m.Provider),
		m.ProviderId,
		"",
	)
	bind.SetId(m.Id)
	if m.ProviderName != nil {
		bind.SetProviderName(*m.ProviderName)
	}
	bind.SetCreatedAt(m.CreatedAt)
	return bind
}

// 确保 ThirdPartyBindRepositoryImpl 实现了 ThirdPartyBindRepository 接口
var _ repository.ThirdPartyBindRepository = (*ThirdPartyBindRepositoryImpl)(nil)
