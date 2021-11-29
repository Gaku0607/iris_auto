package tool

import (
	"os"
	"regexp"
	"strings"
)

func ErrMsgs(assertion bool, msg string) {
	if !assertion {
		panic(msg)
	}
}

//查看字符串是否非數字組成
func IsNumeric(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func IsCsvFormat(filename string) bool {
	b, err := regexp.MatchString(`(csv|CSV){1}`, filename)
	if err != nil {
		return false
	}
	return b
}

func IsExist(path string) bool {

	if _, err := os.Stat(path); err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func GetUniqueCode(origincode string) string {
	origincode = strings.TrimSpace(origincode)
	//取注文番號前14碼
	if (strings.Contains(origincode, "-") || origincode[len(origincode)-1] == 'A') && len(origincode) >= 14 {
		return origincode[:14]
	} else {
		return origincode
	}
}
