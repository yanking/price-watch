package repository

import (
	"context"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
)

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	FindById(ctx context.Context, id int64) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByPhone(ctx context.Context, areaCode, phone string) (*entity.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByPhone(ctx context.Context, areaCode, phone string) (bool, error)
}
