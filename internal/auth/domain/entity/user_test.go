package entity_test

import (
	"testing"
	"time"

	"github.com/yanking/price-watch/internal/auth/domain/entity"
)

func TestNewUser(t *testing.T) {
	pwd, _ := entity.NewPassword("Test1234")
	email, _ := entity.NewEmail("test@example.com")

	user, err := entity.NewUser("testuser", pwd, email)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if user.Username() != "testuser" {
		t.Errorf("Username() = %v, want testuser", user.Username())
	}
	if user.Password() == nil {
		t.Error("Password() should not be nil")
	}
	if user.Email() == nil {
		t.Error("Email() should not be nil")
	}
	if !user.IsActive() {
		t.Error("IsActive() should be true for new user")
	}
}

func TestNewUser_EmptyUsername(t *testing.T) {
	_, err := entity.NewUser("", nil, nil)
	if err == nil {
		t.Error("NewUser() should return error for empty username")
	}
}

func TestUserVerifyEmail(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)

	if user.EmailVerified() {
		t.Error("EmailVerified() should be false initially")
	}

	user.VerifyEmail()
	if !user.EmailVerified() {
		t.Error("EmailVerified() should be true after VerifyEmail()")
	}
}

func TestUserChangePassword(t *testing.T) {
	oldPwd, _ := entity.NewPassword("Test1234")
	user, _ := entity.NewUser("testuser", oldPwd, nil)

	newPwd, _ := entity.NewPassword("NewTest12")
	err := user.ChangePassword("Test1234", newPwd)
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}

	if !user.Password().Verify("NewTest12") {
		t.Error("Password should be updated")
	}
}

func TestUserChangePassword_WrongOld(t *testing.T) {
	oldPwd, _ := entity.NewPassword("Test1234")
	user, _ := entity.NewUser("testuser", oldPwd, nil)

	newPwd, _ := entity.NewPassword("NewTest12")
	err := user.ChangePassword("Wrong123", newPwd)
	if err == nil {
		t.Error("ChangePassword() should return error for wrong old password")
	}
}

func TestUserMaskedPhone(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)
	user.SetAreaCode("86")
	user.SetPhone("13800138000")

	if user.MaskedPhone() != "138****8000" {
		t.Errorf("MaskedPhone() = %v, want 138****8000", user.MaskedPhone())
	}
}

func TestUserFullPhone(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)
	user.SetAreaCode("86")
	user.SetPhone("13800138000")

	if user.FullPhone() != "+8613800138000" {
		t.Errorf("FullPhone() = %v, want +8613800138000", user.FullPhone())
	}
}

func TestUserResetPassword(t *testing.T) {
	oldPwd, _ := entity.NewPassword("Test1234")
	user, _ := entity.NewUser("testuser", oldPwd, nil)

	newPwd, _ := entity.NewPassword("NewTest12")
	user.ResetPassword(newPwd)

	if !user.Password().Verify("NewTest12") {
		t.Error("Password should be reset to new password")
	}
	if user.Password().Verify("Test1234") {
		t.Error("Old password should no longer work")
	}
}

func TestUserUpdateProfile(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)

	err := user.UpdateProfile("newusername", "")
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}

	if user.Username() != "newusername" {
		t.Errorf("Username() = %v, want newusername", user.Username())
	}
}

func TestUserUpdateProfile_EmptyUsername(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)

	err := user.UpdateProfile("", "")
	if err == nil {
		t.Error("UpdateProfile() should return error for empty username")
	}
}

func TestUserUpdatePhone(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)
	user.SetAreaCode("86")
	user.SetPhone("13800138000")
	user.VerifyPhone() // 先验证

	user.UpdatePhone("86", "13900139000")

	if user.Phone() != "13900139000" {
		t.Errorf("Phone() = %v, want 13900139000", user.Phone())
	}
	if user.PhoneVerified() {
		t.Error("PhoneVerified should be false after UpdatePhone")
	}
	if user.FullPhone() != "+8613900139000" {
		t.Errorf("FullPhone() = %v, want +8613900139000", user.FullPhone())
	}
}

func TestUserVerifyPhone(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)

	if user.PhoneVerified() {
		t.Error("PhoneVerified should be false initially")
	}

	user.VerifyPhone()
	if !user.PhoneVerified() {
		t.Error("PhoneVerified should be true after VerifyPhone()")
	}
}

func TestUserActivate(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)
	user.Deactive()
	user.Activate()

	if !user.IsActive() {
		t.Error("User should be active after Activate()")
	}
	if user.Status() != entity.UserStatusActive {
		t.Errorf("Status() = %v, want active", user.Status())
	}
}

func TestUserDeactive(t *testing.T) {
	user, _ := entity.NewUser("testuser", nil, nil)
	user.Deactive()

	if user.IsActive() {
		t.Error("User should not be active after Deactive()")
	}
	if user.Status() != entity.UserStatusInactive {
		t.Errorf("Status() = %v, want inactive", user.Status())
	}
}

func TestNewUserFromData(t *testing.T) {
	now := time.Now()
	user, err := entity.NewUserFromData(
		123, "testuser", "hash123", "test@example.com",
		"86", "13800138000", true, true,
		"https://example.com/avatar.png", "测试昵称",
		"active", "github", "oauth123", now, now,
	)

	if err != nil {
		t.Fatalf("NewUserFromData() error = %v", err)
	}

	if user.ID() != 123 {
		t.Errorf("ID() = %v, want 123", user.ID())
	}
	if user.Username() != "testuser" {
		t.Errorf("Username() = %v, want testuser", user.Username())
	}
	if user.Password() == nil {
		t.Error("Password() should not be nil")
	}
	if user.Email() == nil {
		t.Error("Email() should not be nil")
	}
	if user.AreaCode() != "86" {
		t.Errorf("AreaCode() = %v, want 86", user.AreaCode())
	}
	if user.Phone() != "13800138000" {
		t.Errorf("Phone() = %v, want 13800138000", user.Phone())
	}
	if !user.EmailVerified() {
		t.Error("EmailVerified should be true")
	}
	if !user.PhoneVerified() {
		t.Error("PhoneVerified should be true")
	}
	if user.Status() != entity.UserStatusActive {
		t.Errorf("Status() = %v, want active", user.Status())
	}
	if user.OAuthProvider() != "github" {
		t.Errorf("OAuthProvider() = %v, want github", user.OAuthProvider())
	}
	if user.OAuthID() != "oauth123" {
		t.Errorf("OAuthID() = %v, want oauth123", user.OAuthID())
	}
}

func TestNewUserFromData_InvalidEmail(t *testing.T) {
	now := time.Now()
	_, err := entity.NewUserFromData(
		123, "testuser", "hash123", "invalid-email",
		"86", "13800138000", true, true,
		"https://example.com/avatar.png", "测试昵称",
		"active", "", "", now, now,
	)

	if err == nil {
		t.Error("NewUserFromData() should return error for invalid email")
	}
}
