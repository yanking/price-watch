package valueobject

import (
	"testing"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"有效邮箱", "test@example.com", false},
		{"无效邮箱-无@", "testexample.com", true},
		{"无效邮箱-无域名", "test@", true},
		{"空字符串", "", false}, // 可选字段
		{"带子域名", "user@mail.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && email != nil && email.Value() != tt.input {
				t.Errorf("Email.Value() = %v, want %v", email.Value(), tt.input)
			}
		})
	}
}

func TestEmailMask(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	if email.Mask() != "t***@example.com" {
		t.Errorf("Email.Mask() = %v, want t***@example.com", email.Mask())
	}
}
