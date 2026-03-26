package repository

import (
	"context"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/repository"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
	binddao "github.com/yanking/price-watch/internal/auth/infrastructure/persistence/dao"
)

// ThirdPartyBindRepositoryImpl ThirdPartyBindRepository 的实现
// 对于简单场景，仓储直接委托给 DAO
type ThirdPartyBindRepositoryImpl struct {
	bindDAO binddao.ThirdPartyBindDAO
}

// NewThirdPartyBindRepositoryImpl 创建 ThirdPartyBindRepositoryImpl 实例
func NewThirdPartyBindRepositoryImpl(dao binddao.ThirdPartyBindDAO) *ThirdPartyBindRepositoryImpl {
	return &ThirdPartyBindRepositoryImpl{
		bindDAO: dao,
	}
}

// Save 保存第三方绑定（委托给 DAO 的 Insert 方法）
func (r *ThirdPartyBindRepositoryImpl) Save(ctx context.Context, bind *entity.ThirdPartyBind) error {
	return r.bindDAO.Insert(ctx, bind)
}

// Delete 删除第三方绑定（委托给 DAO）
func (r *ThirdPartyBindRepositoryImpl) Delete(ctx context.Context, userId int64, provider valueobject.OAuthProvider) error {
	return r.bindDAO.Delete(ctx, userId, provider)
}

// FindByProvider 根据第三方提供商和提供商ID查找绑定（委托给 DAO）
func (r *ThirdPartyBindRepositoryImpl) FindByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (*entity.ThirdPartyBind, error) {
	return r.bindDAO.FindByProvider(ctx, provider, providerId)
}

// FindByUserId 根据用户ID查找所有第三方绑定（委托给 DAO）
func (r *ThirdPartyBindRepositoryImpl) FindByUserId(ctx context.Context, userId int64) ([]*entity.ThirdPartyBind, error) {
	return r.bindDAO.FindByUserId(ctx, userId)
}

// ExistsByProvider 检查第三方绑定是否已存在（委托给 DAO）
func (r *ThirdPartyBindRepositoryImpl) ExistsByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (bool, error) {
	return r.bindDAO.ExistsByProvider(ctx, provider, providerId)
}

// 确保 ThirdPartyBindRepositoryImpl 实现了 ThirdPartyBindRepository 接口
var _ repository.ThirdPartyBindRepository = (*ThirdPartyBindRepositoryImpl)(nil)
