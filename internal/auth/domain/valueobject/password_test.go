package valueobject_test

import (
	"testing"
	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"有效密码", "Test1234", false},
		{"太短", "Test12", true},
		{"太长", "Test12345678901234567890", true},
		{"无数字", "Testtest", true},
		{"无字母", "12345678", true},
		{"空字符串", "", false}, // 可选字段，第三方登录无密码
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwd, err := valueobject.NewPassword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && pwd != nil && pwd.Hash() == "" {
				t.Error("Password.Hash() should not be empty")
			}
		})
	}
}

func TestPasswordVerify(t *testing.T) {
	pwd, _ := valueobject.NewPassword("Test1234")

	if !pwd.Verify("Test1234") {
		t.Error("Verify() should return true for correct password")
	}
	if pwd.Verify("Wrong123") {
		t.Error("Verify() should return false for wrong password")
	}
}

func TestNewPasswordFromHash(t *testing.T) {
	pwd1, _ := valueobject.NewPassword("Test1234")
	pwd2 := valueobject.NewPasswordFromHash(pwd1.Hash())

	if !pwd2.Verify("Test1234") {
		t.Error("NewPasswordFromHash should preserve verification")
	}
}
