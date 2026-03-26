package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

type Email struct {
	value string
}

func NewEmail(value string) (*Email, error) {
	if value == "" {
		return nil, nil
	}

	// 去除前后空格
	value = strings.TrimSpace(value)

	// 再次检查是否为空字符串（处理纯空格输入）
	if value == "" {
		return nil, nil
	}

	// 验证邮箱格式
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		return nil, errors.New("邮箱格式不正确")
	}

	return &Email{value: value}, nil
}

func (e *Email) Value() string {
	return e.value
}

func (e *Email) Mask() string {
	if e.value == "" {
		return ""
	}
	at := strings.Index(e.value, "@")
	if at <= 1 {
		return e.value
	}
	return string(e.value[0]) + "***" + e.value[at:]
}
