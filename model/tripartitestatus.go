package model

const (
	TRIPARTITE_BACK_AND_FORTH = "來回件"
	TRIPARTITE_STATUS_NULL    = "NULL"
	TRITARTITE_RE_SEND        = "補寄商品"
)

func IsTripartiteStatus(status string) bool {
	return status == TRIPARTITE_BACK_AND_FORTH || status == TRITARTITE_RE_SEND
}
