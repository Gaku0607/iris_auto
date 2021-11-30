package dividewarehouse

import (
	"sort"
	"strconv"
	"time"

	"github.com/Gaku0607/excelgo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/model"
)

//宅配細項
type DeliveryDetail struct {
	Num           int
	DeliveryOrder string //宅配單號
	ItemCode      string //商品碼
	CountStr      string //箱數大於1時需顯示的欄位
	BoxCount      int    //箱數
	AriaCode      string //區域代碼
	Place         string //地點
	Aria          string //倉庫區域
}

//返回 序號 客代 件數 箱數 區域代碼（宅配） 地點 倉別
func (d *DeliveryDetail) GetWendaRow() []interface{} {
	return []interface{}{d.Num, model.Environment.IDS.WendaMergeBox.CustomerCode + time.Now().Format("20060102"), d.CountStr, d.BoxCount, d.AriaCode, d.Place, d.Aria}
}

//返回 序號 訂單編號-1 箱數 倉別 尾碼
func (d *DeliveryDetail) GetZhaipeiRow() []interface{} {
	return []interface{}{d.Num, d.DeliveryOrder, d.BoxCount, d.Aria, d.ItemCode}
}

type ItemPositionList []*ItemPosition

//將商品進行分區
func (d *ItemPositionList) Divide(itemorders MultipleItemOrders, dds []*DeliveryDetail, sf *excelgo.Sourc) error {
	var obj *ItemPosition

	for _, item := range itemorders {

		var formidx int = -1
		var formarea string

		if obj = d.returnObj(item.ItemCode); obj == nil {
			for _, idx := range item.OrderForms.GetAllOrdersIndex() {
				dds[idx].Aria = iris.UNKNOW_ITEM
			}
			continue
			// return fmt.Errorf("input ItemCode:%s, No items found", item.ItemCode)
		}

		//確認該商品的整體數量
		// if !obj.checkItemCount(item.OrderForms.GetTotal()) {
		// 	return errors.New("Incorrect quantity")
		// }

		//如果為單倉返回實體
		if area, b := obj.GetSingleWareHouse(); b {
			if err := obj.distributionRemainder(area, item.OrderForms, dds); err != nil {
				return err
			}
			continue
		}

		//查看有無不用判斷的分倉的數量
		for idx, form := range item.OrderForms {
			if area, b := obj.DivideWareHouseByItemCount(form.GetTotal()); b {
				formarea = area
				formidx = idx
				break
			}
		}

		if formidx != -1 {

			otherarea := getOtherArea(formarea)

			for _, idx := range item.OrderForms[formidx].GetOrderIndex() {
				dds[idx].Aria = formarea
			}

			if err := obj.distributionRemainder(otherarea, item.OrderForms.DeleteContent(formidx).(OrderForms), dds); err != nil {
				return err
			}
			continue
		}

		//進行複雜分倉 以N2計算
		if err := obj.Divide(item.OrderForms); err != nil {
			return err
		}

		//設置ＱＣ單的Detail
		obj.setDeliveryDetail(dds, item.OrderForms)

	}

	return nil
}

func (l ItemPositionList) returnObj(code string) *ItemPosition {
	for _, obj := range l {
		if obj.ItemCode == code {
			return obj
		}
	}
	return nil
}

//分倉
type ItemPosition struct {
	ItemCode      string
	IsWarehouses  bool           //是否有多倉
	WarehouseInfo map[string]int // 倉別 數量
}

//獲取各個商品的各倉別資訊（數量）
func NewDivideWarehouses(csvsourc *excelgo.Sourc) (ItemPositionList, error) {

	var positionlist []*ItemPosition

	csvparms := model.Environment.IDS.CsvSourc

	areacol := csvsourc.GetCol(csvparms.AreaSpan)
	totalcol := csvsourc.GetCol(csvparms.ItemsTotalSpan)
	itemcodecol := csvsourc.GetCol(csvparms.CodeSpan)

	for _, rows := range csvsourc.GetRows() {

		var flag bool
		itemcode := rows[itemcodecol.Col]
		area := iris.GetAria(rows[areacol.Col])

		total, err := strconv.Atoi(rows[totalcol.Col])
		if err != nil {
			return nil, err
		}

		for _, pl := range positionlist {
			if pl.isSame(itemcode) {
				pl.MergeWareHouse(itemcode, area, total)
				flag = true
			}
		}

		if !flag {
			sw := &ItemPosition{ItemCode: itemcode, WarehouseInfo: map[string]int{area: total}}
			positionlist = append(positionlist, sw)
		}

	}
	return ItemPositionList(positionlist), nil
}

//將同倉別的物品累計數量 不同的倉別則新增
func (pl *ItemPosition) MergeWareHouse(code, area string, total int) {
	if val, exit := pl.WarehouseInfo[area]; exit {
		pl.WarehouseInfo[area] = val + total
	} else {
		pl.WarehouseInfo[area] = total
		pl.IsWarehouses = true
	}
}

//查詢有無指定數量的倉庫
func (pl *ItemPosition) DivideWareHouseByItemCount(count int) (string, bool) {
	for area, total := range pl.WarehouseInfo {
		if total == count {
			return area, true
		}
	}
	return "", false
}

//獲取單倉資訊 如為多倉false
func (pl *ItemPosition) GetSingleWareHouse() (string, bool) {
	if !pl.IsWarehouses {
		for Id := range pl.WarehouseInfo {
			return Id, true
		}
	}
	return "", false
}

//商品在所有倉庫的總數
func (pl *ItemPosition) itemTotal() int {
	var total int
	for _, v := range pl.WarehouseInfo {
		total += v
	}
	return total
}

//確認
func (pl *ItemPosition) isSame(itemcode string) bool {
	return pl.ItemCode == itemcode
}

//確認商品總數
func (pl *ItemPosition) checkItemCount(count int) bool {
	return pl.itemTotal() >= count
}

//把相同商品碼的檔案進行分倉
func (pl *ItemPosition) Divide(orders OrderForms) error {

	//只計算N2倉
	count, _ := pl.WarehouseInfo[iris.NEXT2]

	//該商品請求總數小於Ｎ2倉庫數量時
	if orders.GetTotal() <= count {
		orders.SetAllArea(iris.NEXT2)
		return nil
	}

	idx := len(orders)
	sort.Sort(orders)

	// idx為-1時為 完整的分配了貨物 無任何單品項多倉問題多倉
	// idx為0時為 Order的最小請求數量有可能超過Ｎ2倉
	// idx大於0時為 無法以該idx的數量 完整分配仍須重新分配

	for {
		if idx = pl.sub(count, orders[:idx]); idx != -1 {
			if idx == 0 && orders[idx].GetTotal() > count {
				pl.divide(count, orders[0])
				break
			}
			if idx == 0 && orders[idx].GetTotal() < count {
				idx++
			}

			//當 Ｎ2倉 無法把商品分配至0品項時 將會出現一項出貨指示有多倉
			if !pl.isEnough(orders[:idx], count) {
				for i := range orders[:idx] {
					orders[i].SetAllArea(iris.NEXT2)
				}
				//設置為多倉
				pl.divide(count-orders[:idx].GetTotal(), orders[idx])
				break
			}

		} else {
			break
		}
	}

	count, _ = pl.WarehouseInfo[iris.NEXT3]

	//剩餘的帶入Ｎ3倉
	for _, order := range orders {

		for idx, rowinfo := range order.Info {
			if rowinfo.area == "" {

				if count <= 0 {
					order.Info[idx].area = iris.UNKNOW
					continue
				}

				count -= rowinfo.count

				order.Info[idx].area = iris.IsUnKnowWareHouse(count, iris.NEXT3)

			}
		}
	}

	return nil
}

//計算剩餘數量
func (pl *ItemPosition) distributionRemainder(area string, orders OrderForms, dds []*DeliveryDetail) error {

	count, _ := pl.WarehouseInfo[area]

	if orders.GetTotal() <= count {
		orders.SetAllArea(area)
		pl.setDeliveryDetail(dds, orders)
		return nil
	}

	sort.Sort(orders)

	for _, order := range orders {

		if count <= 0 {
			order.SetAllArea(iris.UNKNOW)
			continue
		}

		tempcount := count - order.GetTotal()

		if tempcount >= 0 {
			order.SetAllArea(area)
			count -= order.GetTotal()
		} else {

			//當剩餘數量配不滿且只有一則訂單時
			if order.Len() == 1 {
				order.SetAllArea(iris.UnKnowWareHouse(area))
				count = tempcount
				continue
			}

			sort.Sort(order)

			for i := range order.Info {

				if count <= 0 {
					order.SetArea(i, iris.UNKNOW)
					continue
				}

				count -= order.GetCount(i)

				order.SetArea(i, iris.IsUnKnowWareHouse(count, area))
			}

		}
	}

	pl.setDeliveryDetail(dds, orders)

	return nil
}

//已給個row為單位再細分貨品
func (pl *ItemPosition) divide(count int, orders *OrderForm) {

	if count == 0 {
		return
	}

	if orders.Len() == 1 {
		pl.WarehouseInfo[iris.NEXT3] -= (orders.GetTotal() - count)
		orders.SetArea(
			0, iris.IsUnKnowWareHouse(pl.WarehouseInfo[iris.NEXT3], iris.MIXNEXT),
		)
		return
	}
	// idx為-1時為 完整的分配了貨物 無任何單品項多倉問題多倉
	// idx為0時為 Order的最小請求數量有可能超過Ｎ2倉
	// idx大於0時為 無法以該idx的數量 完整分配仍須重新分配

	sort.Sort(orders)
	var idx int
	for {
		if idx = pl.sub(count, orders); idx != -1 {

			if idx == 0 && orders.Info[idx].count > count {
				pl.WarehouseInfo[iris.NEXT3] -= (orders.Info[idx].count - count) //減去拿取Ｎ3的量
				orders.SetArea(
					0, iris.IsUnKnowWareHouse(pl.WarehouseInfo[iris.NEXT3], iris.MIXNEXT),
				)
				return
			}

			if idx == 0 && orders.Info[idx].count < count { //當idx為0有可能是只有0滿足條件
				idx++
			}

			temp := orders.Info
			orders.Info = orders.Info[:idx]

			if !pl.isEnough(orders, count) {

				orders.Info = temp
				sum := 0

				for i := range orders.Info[:idx] {
					orders.SetArea(i, iris.NEXT2)
					sum += orders.GetCount(i)
				}

				//減去拿取Ｎ3的量
				pl.WarehouseInfo[iris.NEXT3] -= orders.Info[idx].count - (count - sum)
				orders.SetArea(
					0, iris.IsUnKnowWareHouse(pl.WarehouseInfo[iris.NEXT3], iris.MIXNEXT),
				)
				return
			}

		} else {
			return
		}
	}
}

func (pl *ItemPosition) setDeliveryDetail(dds []*DeliveryDetail, orders OrderForms) {
	for idx, area := range orders.GetAllOrdersInfo() {
		dds[idx].Aria = area
	}
	for idx := range dds {
		if dds[idx].Aria == "" {
			dds[idx].Aria = iris.UNCONFIRMED
		}
	}
}

func (pl *ItemPosition) sub(count int, orders Orders) int {
	max, index := pl.find(0, orders.Len()-1, orders.Len()/2, count, orders)
	if count-max < 0 {
		return index
	} else if count-max > 0 {
		//使用append底層仍然是用同一個數組 需要使用ＣＯＰＹ
		neworders := orders.DeleteContent(index)

		if flag := pl.sub(count-max, neworders); flag == -1 { //flag正確為0時將會紀錄 item的倉別
			orders.SetArea(index, iris.NEXT2)
			return flag
		} else {
			return index
		}
	} else { //count-max == 0
		orders.SetArea(index, iris.NEXT2)
		return -1
	}
}

//返回Target 如果沒有Target返回離Target最接近的小於數,以及座標
func (pl *ItemPosition) find(left, right, mid, target int, orders Orders) (int, int) {
	if orders.GetCount(mid) == target {
		return target, mid
	} else if left >= right {
		if orders.GetCount(mid) > target && mid-1 >= 0 { //此時mid有可能為0
			return orders.GetCount(mid - 1), mid - 1
		}
		return orders.GetCount(mid), mid
	} else if orders.GetCount(mid) > target {
		return pl.find(left, mid-1, (left+mid-1)/2, target, orders)
	} else if orders.GetCount(mid) < target {
		return pl.find(mid+1, right, (right+mid+1)/2, target, orders)
	}
	return -1, -1
}

func (pl *ItemPosition) isEnough(orders Orders, count int) bool {
	return count <= orders.GetTotal()
}

func getOtherArea(mainarea string) string {
	if mainarea == iris.NEXT2 {
		return iris.NEXT3
	} else {
		return iris.NEXT2
	}
}
