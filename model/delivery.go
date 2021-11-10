package model

//讀取config獲取各個郵遞區域
var DeliveryArea map[string]*Delivery = map[string]*Delivery{}

const (
	MORNING   = "早上"
	AFTERNOON = "下午"
	OTHER     = "other"
)

type Delivery struct {
	City       string `json:"city"`        //城市
	Partition  string `json:"partition"`   //區域
	AreaCode   string `json:"aria_code"`   //區域代碼
	PostalCode string `json:"postal_code"` //郵遞區號
	Place      string `json:"place"`       //地點
}

type WarehouseItemInfo struct {
	Index       int
	Data        []interface{}
	WarehouseID string
}
