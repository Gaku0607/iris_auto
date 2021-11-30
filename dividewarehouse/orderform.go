package dividewarehouse

import (
	"sort"
)

type Orders interface {
	sort.Interface
	GetTotal() int
	GetCount(int) int
	SetArea(int, string)
	SetAllArea(string)
	DeleteContent(int) Orders
}

type OrderForms []*OrderForm

func (f OrderForms) IsOrderExist(ordercode string) (*OrderForm, bool) {
	for _, order := range f {
		if order.OrderCode == ordercode {
			return order, true
		}
	}
	return nil, false
}

func (f OrderForms) GetAllOrdersIndex() []int {
	var list []int
	for _, order := range f {
		list = append(list, order.GetOrderIndex()...)
	}
	return list
}

func (f OrderForms) GetAllOrdersInfo() map[int]string {
	var info map[int]string = make(map[int]string)
	for _, order := range f {
		for _, rowinfo := range order.Info {
			info[rowinfo.index] = rowinfo.area
		}
	}
	return info
}

//實現Order
func (f OrderForms) GetTotal() int {
	var itemtotal int
	for _, order := range f {
		itemtotal += order.GetTotal()
	}
	return itemtotal
}

func (f OrderForms) SetAllArea(area string) {
	for idx := range f {
		f[idx].SetAllArea(area)
	}
}

func (f OrderForms) SetArea(idx int, area string) {
	for i := range f[idx].Info {
		f[idx].Info[i].area = area
	}
}

func (f OrderForms) GetCount(idx int) int {
	return f[idx].Total
}

func (f OrderForms) DeleteContent(idx int) Orders {
	var neworders OrderForms = make(OrderForms, len(f)-1)
	copy(neworders, f[:idx])
	copy(neworders[len(f[:idx]):], f[idx+1:])
	return neworders
}

//實現Sort
func (f OrderForms) Len() int           { return len(f) }
func (f OrderForms) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f OrderForms) Less(i, j int) bool { return f[i].Total < f[j].Total }

type rowinfo struct {
	index int
	count int
	area  string
}

//商品訂單
type OrderForm struct {
	OrderCode string
	Info      []*rowinfo //原座標以及倉別
	Total     int
}

func NewOrderForm(ordercode string, count, index int) *OrderForm {
	return &OrderForm{OrderCode: ordercode, Total: count, Info: []*rowinfo{{index: index, count: count}}}
}

//實現Sort
func (of *OrderForm) Len() int           { return len(of.Info) }
func (of *OrderForm) Less(i, j int) bool { return of.Info[i].count < of.Info[j].count }
func (of *OrderForm) Swap(i, j int) {
	of.Info[i], of.Info[j] = of.Info[j], of.Info[i]
}

func (of *OrderForm) GetTotal() int {
	sum := 0
	for _, rowinfo := range of.Info {
		sum += rowinfo.count
	}
	return sum
}
func (of *OrderForm) GetCount(idx int) int         { return of.Info[idx].count }
func (of *OrderForm) SetArea(idx int, area string) { of.Info[idx].area = area }
func (of *OrderForm) DeleteContent(idx int) Orders {
	var newInfo []*rowinfo = make([]*rowinfo, of.Len()-1)
	copy(newInfo, of.Info[:idx])
	copy(newInfo[len(of.Info[:idx]):], of.Info[idx+1:])
	return &OrderForm{
		OrderCode: of.OrderCode,
		Total:     of.Total,
		Info:      newInfo,
	}
}

func (of *OrderForm) SetAllArea(area string) {
	for i := range of.Info {
		of.Info[i].area = area
	}
}

func (of *OrderForm) addOrderInfo(count, index int) {
	of.Info = append(of.Info, &rowinfo{index: index, count: count})
	of.Total += count
}

func (of *OrderForm) GetOrderIndex() []int {
	var (
		list []int = make([]int, len(of.Info))
	)
	for idx, val := range of.Info {
		list[idx] = val.index
	}
	return list
}
