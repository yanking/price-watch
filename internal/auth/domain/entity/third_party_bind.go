package entity

import (
	"time"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

type ThirdPartyBind struct {
	id           int64
	userId       int64
	provider     valueobject.OAuthProvider
	providerId   string
	providerName string
	createdAt    time.Time
}

func NewThirdPartyBind(userId int64, provider valueobject.OAuthProvider, providerId, providerName string) *ThirdPartyBind {
	return &ThirdPartyBind{
		userId:       userId,
		provider:     provider,
		providerId:   providerId,
		providerName: providerName,
		createdAt:    time.Now(),
	}
}

// Getters
func (b *ThirdPartyBind) Id() int64                          { return b.id }
func (b *ThirdPartyBind) UserId() int64                      { return b.userId }
func (b *ThirdPartyBind) Provider() valueobject.OAuthProvider { return b.provider }
func (b *ThirdPartyBind) ProviderId() string                 { return b.providerId }
func (b *ThirdPartyBind) ProviderName() string               { return b.providerName }
func (b *ThirdPartyBind) CreatedAt() time.Time               { return b.createdAt }

// Setters
func (b *ThirdPartyBind) SetId(id int64)                     { b.id = id }
func (b *ThirdPartyBind) SetProviderName(name string)        { b.providerName = name }
func (b *ThirdPartyBind) SetCreatedAt(t time.Time)           { b.createdAt = t }
