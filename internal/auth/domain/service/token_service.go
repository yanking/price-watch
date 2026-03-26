package service

import "github.com/yanking/price-watch/internal/auth/domain/entity"

type TokenService interface {
    GenerateToken(user *entity.User) (token string, version int64, err error)
    ParseToken(token string) (userId int64, err error)
    IncrementVersion(userId int64) (int64, error)
    GetVersion(userId int64) (int64, error)
}
