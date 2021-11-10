package iris

import "strings"

const (
	Next2   = "N2"
	Next3   = "N3"
	MixNext = Next2 + "." + Next3
	NA      = "N/A"
)

//判斷倉庫位置
func GetAria(area string) string {
	if strings.Contains(area, "N") {
		return Next3
	} else {
		if area != "" {
			return Next2
		} else {
			return NA
		}
	}
}
