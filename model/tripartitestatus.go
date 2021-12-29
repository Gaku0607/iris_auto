package model

const (
	// TRIPARTITE_BACK_AND_FORTH = "來回件"
	// TRITARTITE_RE_SEND        = "補寄商品"
	// TRITARTITE_MOMO_CHANGE    = "momo第三方交換"
	TRIPARTITE_STATUS_NULL = "NULL"
)

var TripartiteStatus []string //第三方ＱＣ製作中需要的狀態

func IsTripartiteStatus(status string) bool {
	for _, s := range TripartiteStatus {
		if s == status {
			return true
		}
	}
	return false
}
