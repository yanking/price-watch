package valueobject_test

import (
	"testing"

	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestOAuthProviderString(t *testing.T) {
	tests := []struct {
		provider valueobject.OAuthProvider
		want     string
	}{
		{valueobject.OAuthProviderWeChat, "wechat"},
		{valueobject.OAuthProviderGitHub, "github"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.provider.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOAuthProvider(t *testing.T) {
	tests := []struct {
		input   string
		want    valueobject.OAuthProvider
		wantErr bool
	}{
		{"wechat", valueobject.OAuthProviderWeChat, false},
		{"github", valueobject.OAuthProviderGitHub, false},
		{"unknown", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := valueobject.ParseOAuthProvider(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOAuthProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseOAuthProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}
