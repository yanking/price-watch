package entity_test

import (
	"testing"
	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewThirdPartyBind(t *testing.T) {
	bind := entity.NewThirdPartyBind(1, valueobject.OAuthProviderGitHub, "12345", "octocat")

	if bind.UserId() != 1 {
		t.Errorf("UserId() = %v, want 1", bind.UserId())
	}
	if bind.Provider() != valueobject.OAuthProviderGitHub {
		t.Errorf("Provider() = %v, want github", bind.Provider())
	}
	if bind.ProviderId() != "12345" {
		t.Errorf("ProviderId() = %v, want 12345", bind.ProviderId())
	}
	if bind.ProviderName() != "octocat" {
		t.Errorf("ProviderName() = %v, want octocat", bind.ProviderName())
	}
}
