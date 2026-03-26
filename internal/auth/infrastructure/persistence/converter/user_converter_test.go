package converter_test

import (
	"testing"
	"time"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
	"github.com/yanking/price-watch/internal/auth/infrastructure/persistence/converter"
	"github.com/yanking/price-watch/internal/auth/infrastructure/persistence/mysql"
)

func TestUserToPO(t *testing.T) {
	// 准备测试数据
	pwd, _ := valueobject.NewPassword("Test1234")
	email, _ := valueobject.NewEmail("test@example.com")

	user, err := entity.NewUser("testuser", pwd, email)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 设置一些额外字段
	user.SetID(1234567890)
	user.SetAreaCode("86")
	user.SetPhone("13800138000")

	// DO -> PO
	po := converter.UserToPO(user)

	// 验证基本字段
	if po.Id != int64(1234567890) {
		t.Errorf("Id = %v, want 1234567890", po.Id)
	}
	if po.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", po.Username)
	}

	// 验证密码哈希
	if po.PasswordHash == nil {
		t.Error("PasswordHash should not be nil")
	} else {
		// 验证密码哈希可以正确验证原密码
		pwdFromHash := valueobject.NewPasswordFromHash(*po.PasswordHash)
		if !pwdFromHash.Verify("Test1234") {
			t.Error("Password hash verification failed")
		}
	}

	// 验证邮箱
	if po.Email == nil {
		t.Error("Email should not be nil")
	} else if *po.Email != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", *po.Email)
	}

	// 验证区号
	if po.AreaCode == nil {
		t.Error("AreaCode should not be nil")
	} else if *po.AreaCode != "86" {
		t.Errorf("AreaCode = %v, want 86", *po.AreaCode)
	}

	// 验证手机号
	if po.Phone == nil {
		t.Error("Phone should not be nil")
	} else if *po.Phone != "13800138000" {
		t.Errorf("Phone = %v, want 13800138000", *po.Phone)
	}

	// 验证状态
	if po.Status != 1 {
		t.Errorf("Status = %v, want 1 (active)", po.Status)
	}

	// 验证时间
	if !po.CreatedAt.IsZero() && !user.CreatedAt().Equal(po.CreatedAt) {
		t.Error("CreatedAt not preserved")
	}
}

func TestPOToUser(t *testing.T) {
	// 准备 PO 数据
	now := time.Now()
	hash, _ := valueobject.NewPassword("Test1234")
	email := "test@example.com"
	areaCode := "86"
	phone := "13800138000"

	po := &mysql.UserPO{
		Id:            1234567890,
		Username:      "testuser",
		PasswordHash:  func() *string { h := hash.Hash(); return &h }(),
		Email:         &email,
		EmailVerified: true,
		AreaCode:      &areaCode,
		Phone:         &phone,
		PhoneVerified: false,
		Status:        1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// PO -> DO
	user := converter.POToUser(po)

	// 验证基本字段
	if user.ID() != uint64(1234567890) {
		t.Errorf("ID = %v, want 1234567890", user.ID())
	}
	if user.Username() != "testuser" {
		t.Errorf("Username = %v, want testuser", user.Username())
	}

	// 验证密码
	if user.Password() == nil {
		t.Error("Password should not be nil")
	} else if !user.Password().Verify("Test1234") {
		t.Error("Password verification failed after conversion")
	}

	// 验证邮箱
	if user.Email() == nil {
		t.Error("Email should not be nil")
	} else if user.Email().Value() != "test@example.com" {
		t.Errorf("Email = %v, want test@example.com", user.Email().Value())
	}

	// 验证区号
	if user.AreaCode() != "86" {
		t.Errorf("AreaCode = %v, want 86", user.AreaCode())
	}

	// 验证手机号
	if user.Phone() != "13800138000" {
		t.Errorf("Phone = %v, want 13800138000", user.Phone())
	}

	// 验证邮箱验证状态
	if user.EmailVerified() != true {
		t.Error("EmailVerified should be true")
	}

	// 验证手机验证状态
	if user.PhoneVerified() != false {
		t.Error("PhoneVerified should be false")
	}

	// 验证状态
	if user.Status() != entity.UserStatusActive {
		t.Errorf("Status = %v, want active", user.Status())
	}
}

func TestUserConverterRoundTrip(t *testing.T) {
	// 准备测试数据
	pwd, _ := valueobject.NewPassword("Test1234")
	email, _ := valueobject.NewEmail("test@example.com")

	user1, err := entity.NewUser("testuser", pwd, email)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user1.SetID(1234567890)
	user1.SetAreaCode("86")
	user1.SetPhone("13800138000")

	// DO -> PO -> DO
	po := converter.UserToPO(user1)
	user2 := converter.POToUser(po)

	// 验证往返转换后数据一致性
	if user1.ID() != user2.ID() {
		t.Errorf("ID not preserved: %v -> %v", user1.ID(), user2.ID())
	}
	if user1.Username() != user2.Username() {
		t.Errorf("Username not preserved: %v -> %v", user1.Username(), user2.Username())
	}
	if user1.AreaCode() != user2.AreaCode() {
		t.Errorf("AreaCode not preserved: %v -> %v", user1.AreaCode(), user2.AreaCode())
	}
	if user1.Phone() != user2.Phone() {
		t.Errorf("Phone not preserved: %v -> %v", user1.Phone(), user2.Phone())
	}
	if user1.EmailVerified() != user2.EmailVerified() {
		t.Errorf("EmailVerified not preserved: %v -> %v", user1.EmailVerified(), user2.EmailVerified())
	}
	if user1.PhoneVerified() != user2.PhoneVerified() {
		t.Errorf("PhoneVerified not preserved: %v -> %v", user1.PhoneVerified(), user2.PhoneVerified())
	}
	if user1.Status() != user2.Status() {
		t.Errorf("Status not preserved: %v -> %v", user1.Status(), user2.Status())
	}

	// 验证密码可以正确验证
	if user2.Password() == nil {
		t.Error("Password should not be nil after round trip")
	} else if !user2.Password().Verify("Test1234") {
		t.Error("Password verification failed after round trip")
	}

	// 验证邮箱
	if user2.Email() == nil {
		t.Error("Email should not be nil after round trip")
	} else if user2.Email().Value() != "test@example.com" {
		t.Errorf("Email not preserved: %v -> %v", user1.Email().Value(), user2.Email().Value())
	}
}

func TestThirdPartyBindToPO(t *testing.T) {
	bind := entity.NewThirdPartyBind(
		1234567890,
		valueobject.OAuthProviderGitHub,
		"github_user_123",
		"GitHub User",
	)
	bind.SetId(987654321)

	po := converter.ThirdPartyBindToPO(bind)

	if po.Id != int64(987654321) {
		t.Errorf("Id = %v, want 987654321", po.Id)
	}
	if po.UserId != int64(1234567890) {
		t.Errorf("UserId = %v, want 1234567890", po.UserId)
	}
	if po.Provider != int8(valueobject.OAuthProviderGitHub) {
		t.Errorf("Provider = %v, want %v", po.Provider, valueobject.OAuthProviderGitHub)
	}
	if po.ProviderId != "github_user_123" {
		t.Errorf("ProviderId = %v, want github_user_123", po.ProviderId)
	}
	if po.ProviderName == nil {
		t.Error("ProviderName should not be nil")
	} else if *po.ProviderName != "GitHub User" {
		t.Errorf("ProviderName = %v, want GitHub User", *po.ProviderName)
	}
}

func TestPOToThirdPartyBind(t *testing.T) {
	now := time.Now()
	providerName := "GitHub User"

	po := &mysql.ThirdPartyBindPO{
		Id:           987654321,
		UserId:       1234567890,
		Provider:     int8(valueobject.OAuthProviderGitHub),
		ProviderId:   "github_user_123",
		ProviderName: &providerName,
		CreatedAt:    now,
	}

	bind := converter.POToThirdPartyBind(po)

	if bind.Id() != int64(987654321) {
		t.Errorf("Id = %v, want 987654321", bind.Id())
	}
	if bind.UserId() != int64(1234567890) {
		t.Errorf("UserId = %v, want 1234567890", bind.UserId())
	}
	if bind.Provider() != valueobject.OAuthProviderGitHub {
		t.Errorf("Provider = %v, want %v", bind.Provider(), valueobject.OAuthProviderGitHub)
	}
	if bind.ProviderId() != "github_user_123" {
		t.Errorf("ProviderId = %v, want github_user_123", bind.ProviderId())
	}
	if bind.ProviderName() != "GitHub User" {
		t.Errorf("ProviderName = %v, want GitHub User", bind.ProviderName())
	}
}

func TestUserConverterWithNilFields(t *testing.T) {
	// 测试所有可选字段都为 nil 的情况
	user, err := entity.NewUser("minimal_user", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	user.SetID(123)

	po := converter.UserToPO(user)

	if po.PasswordHash != nil {
		t.Error("PasswordHash should be nil")
	}
	if po.Email != nil {
		t.Error("Email should be nil")
	}
	if po.AreaCode != nil {
		t.Error("AreaCode should be nil")
	}
	if po.Phone != nil {
		t.Error("Phone should be nil")
	}
	if po.Avatar != nil {
		t.Error("Avatar should be nil")
	}
	if po.Nickname != nil {
		t.Error("Nickname should be nil")
	}
}
