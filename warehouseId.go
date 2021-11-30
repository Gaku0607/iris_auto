package iris

import "strings"

const (
	NEXT2       = "N2"                //N2倉庫
	NEXT3       = "N3"                //N3倉庫
	MIXNEXT     = NEXT2 + "." + NEXT3 //混合倉
	UNKNOW      = "N/A"               //未知倉庫（數量不足時產生）
	UNKNOW_ITEM = "UNKNOW"            //未知品項（商品6碼查無時）
	UNCONFIRMED = "UNCONFIRMED"       //待確認
)

//判斷倉庫位置
func GetAria(area string) string {
	if strings.Contains(area, "N") {
		return NEXT3
	} else {
		if area != "" {
			return NEXT2
		} else {
			return UNKNOW
		}
	}
}
func IsUnKnowWareHouse(count int, area string) string {
	if count < 0 {
		if area == MIXNEXT {
			return NEXT2 + "." + "(" + NEXT3 + "." + UNKNOW + ")"
		}

		return "(" + area + "." + UNKNOW + ")"
	} else {
		return area
	}
}

func UnKnowWareHouse(area string) string {
	return "(" + area + "." + UNKNOW + ")"
}
