package entity_test

import (
	"testing"

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
