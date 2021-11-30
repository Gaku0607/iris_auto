package dividewarehouse

//多件商品訂單
type MultipleItemOrders []*ItemOrder

func (m MultipleItemOrders) IsItemExist(itemcode string) (*ItemOrder, bool) {
	for _, order := range m {
		if order.ItemCode == itemcode {
			return order, true
		}
	}
	return nil, false
}

//指定商品的所有訂單
type ItemOrder struct {
	ItemCode   string     //商品碼
	OrderForms OrderForms //商品所有訂單
}

func NewItemOrder(itemcode, ordercode string, count, index int) *ItemOrder {
	return &ItemOrder{ItemCode: itemcode, OrderForms: OrderForms{NewOrderForm(ordercode, count, index)}}
}

func (io *ItemOrder) AddOrderForm(ordercode string, count, index int) {
	if orderform, b := io.OrderForms.IsOrderExist(ordercode); b {
		orderform.addOrderInfo(count, index)
	} else {
		io.OrderForms = append(io.OrderForms, NewOrderForm(ordercode, count, index))
	}
}
