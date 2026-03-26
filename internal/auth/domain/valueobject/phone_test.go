package valueobject_test

import (
	"testing"

	"github.com/yanking/price-watch/internal/auth/domain/valueobject"
)

func TestNewPhone(t *testing.T) {
	tests := []struct {
		name     string
		areaCode string
		number   string
		wantErr  bool
	}{
		{"中国手机号", "86", "13800138000", false},
		{"美国手机号", "1", "2125551234", false},
		{"无效区号", "0", "13800138000", true},
		{"中国无效号码", "86", "12345", true},
		{"空值", "", "", false},
		{"只有区号", "86", "", true},
		{"只有号码", "", "13800138000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := valueobject.NewPhone(tt.areaCode, tt.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPhone() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && phone != nil {
				if phone.AreaCode() != tt.areaCode {
					t.Errorf("AreaCode() = %v, want %v", phone.AreaCode(), tt.areaCode)
				}
				if phone.Number() != tt.number {
					t.Errorf("Number() = %v, want %v", phone.Number(), tt.number)
				}
			}
		})
	}
}

func TestPhoneFull(t *testing.T) {
	phone, _ := valueobject.NewPhone("86", "13800138000")
	if phone.Full() != "+8613800138000" {
		t.Errorf("Full() = %v, want +8613800138000", phone.Full())
	}
}

func TestPhoneMask(t *testing.T) {
	phone, _ := valueobject.NewPhone("86", "13800138000")
	if phone.Mask() != "138****8000" {
		t.Errorf("Mask() = %v, want 138****8000", phone.Mask())
	}
}
