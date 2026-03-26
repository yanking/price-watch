package repository

import (
	"context"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

type ThirdPartyBindRepository interface {
	Save(ctx context.Context, bind *entity.ThirdPartyBind) error
	Delete(ctx context.Context, userId int64, provider valueobject.OAuthProvider) error
	FindByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (*entity.ThirdPartyBind, error)
	FindByUserId(ctx context.Context, userId int64) ([]*entity.ThirdPartyBind, error)
	ExistsByProvider(ctx context.Context, provider valueobject.OAuthProvider, providerId string) (bool, error)
}
