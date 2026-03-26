package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

type Phone struct {
	areaCode string
	number   string
}

func NewPhone(areaCode, number string) (*Phone, error) {
	if areaCode == "" && number == "" {
		return nil, nil
	}

	if areaCode == "" {
		return nil, errors.New("区号不能为空")
	}
	if number == "" {
		return nil, errors.New("手机号不能为空")
	}

	// 清理格式
	areaCode = strings.TrimPrefix(areaCode, "+")
	areaCode = strings.TrimSpace(areaCode)
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")

	// 验证区号:1-4位数字,不以0开头
	if matched, _ := regexp.MatchString(`^[1-9]\d{0,3}$`, areaCode); !matched {
		return nil, errors.New("区号格式不正确")
	}

	// 根据区号验证号码
	switch areaCode {
	case "86": // 中国
		if matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, number); !matched {
			return nil, errors.New("手机号格式不正确")
		}
	case "1": // 美国/加拿大
		if matched, _ := regexp.MatchString(`^\d{10}$`, number); !matched {
			return nil, errors.New("手机号格式不正确")
		}
	default:
		// 其他国家:6-15位数字
		if matched, _ := regexp.MatchString(`^\d{6,15}$`, number); !matched {
			return nil, errors.New("手机号格式不正确")
		}
	}

	return &Phone{areaCode: areaCode, number: number}, nil
}

func (p *Phone) AreaCode() string {
	return p.areaCode
}

func (p *Phone) Number() string {
	return p.number
}

func (p *Phone) Full() string {
	if p.areaCode == "" || p.number == "" {
		return ""
	}
	return "+" + p.areaCode + p.number
}

func (p *Phone) Mask() string {
	if p.number == "" {
		return ""
	}
	if len(p.number) > 7 {
		return p.number[:3] + "****" + p.number[len(p.number)-4:]
	}
	return p.number
}
