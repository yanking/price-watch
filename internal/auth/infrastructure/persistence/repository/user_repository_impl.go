package repository

import (
	"context"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/repository"
	userdao "github.com/yanking/price-watch/internal/auth/infrastructure/persistence/dao"
)

// UserRepositoryImpl UserRepository 的实现
// 对于简单场景，仓储直接委托给 DAO
type UserRepositoryImpl struct {
	userDAO userdao.UserDAO
}

// NewUserRepositoryImpl 创建 UserRepositoryImpl 实例
func NewUserRepositoryImpl(dao userdao.UserDAO) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		userDAO: dao,
	}
}

// Save 保存用户（委托给 DAO 的 Insert 方法）
func (r *UserRepositoryImpl) Save(ctx context.Context, user *entity.User) error {
	return r.userDAO.Insert(ctx, user)
}

// Update 更新用户（委托给 DAO 的 Update 方法）
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	return r.userDAO.Update(ctx, user)
}

// FindById 根据 ID 查找用户（委托给 DAO）
func (r *UserRepositoryImpl) FindById(ctx context.Context, id int64) (*entity.User, error) {
	return r.userDAO.FindById(ctx, id)
}

// FindByUsername 根据用户名查找用户（委托给 DAO）
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	return r.userDAO.FindByUsername(ctx, username)
}

// FindByEmail 根据邮箱查找用户（委托给 DAO）
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return r.userDAO.FindByEmail(ctx, email)
}

// FindByPhone 根据区号和手机号查找用户（委托给 DAO）
func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, areaCode, phone string) (*entity.User, error) {
	return r.userDAO.FindByPhone(ctx, areaCode, phone)
}

// ExistsByUsername 检查用户名是否已存在（委托给 DAO）
func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.userDAO.ExistsByUsername(ctx, username)
}

// ExistsByEmail 检查邮箱是否已存在（委托给 DAO）
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.userDAO.ExistsByEmail(ctx, email)
}

// ExistsByPhone 检查手机号是否已存在（委托给 DAO）
func (r *UserRepositoryImpl) ExistsByPhone(ctx context.Context, areaCode, phone string) (bool, error) {
	return r.userDAO.ExistsByPhone(ctx, areaCode, phone)
}

// 确保 UserRepositoryImpl 实现了 UserRepository 接口
var _ repository.UserRepository = (*UserRepositoryImpl)(nil)
