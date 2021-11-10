package iris

import (
	"strconv"
	"time"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
)

type DividItems []*DivideWarehouse

//將商品進行分區
func (d *DividItems) Sub(rows [][]interface{}, dds []*DeliveryDetail, sf *excelgo.Sourc) error {

	sws := []*DivideWarehouse(*d)
	itemcodecol := sf.GetCol(model.Environment.IDS.ItemCodeSpan)

	for _, sw := range sws {
		//每個商品ＣＯＤＥ的檔案
		var items []*model.WarehouseItemInfo
		var flag bool //確認商品CODE是否以跟換

		for i, row := range rows {
			if sw.ItemCode == row[itemcodecol.Col].(string) {
				items = append(items, &model.WarehouseItemInfo{Index: i, Data: row})
				flag = true
			} else if flag {
				break
			}
		}

		if err := sw.Sub(sf, items); err != nil {
			return err
		}
		for _, item := range items {
			dds[item.Index].Aria = item.WarehouseID
		}
	}

	return nil
}

const (
	CityCol int = iota
	PartitionCol
	PostalCodeCol
	AreaCodeCol
	PlaceCol
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

//分倉
type DivideWarehouse struct {
	ItemCode string

	IsWarehouses bool //是否有多倉

	WarehouseInfo map[string]int // 倉別 數量

	col int
}

//獲取各個商品的各倉別資訊（數量）
func NewDivideWarehouses(csvsourc *excelgo.Sourc) (DividItems, error) {

	var sws []*DivideWarehouse

	csvparms := model.Environment.IDS.CsvSourc

	areacol := csvsourc.GetCol(csvparms.AreaSpan)
	totalcol := csvsourc.GetCol(csvparms.ItemsTotalSpan)
	itemcodecol := csvsourc.GetCol(csvparms.CodeSpan)

	for _, rows := range csvsourc.GetRows() {

		var flag bool
		itemcode := rows[itemcodecol.Col]
		area := GetAria(rows[areacol.Col])

		total, err := strconv.Atoi(rows[totalcol.Col])
		if err != nil {
			return nil, err
		}

		for _, sw := range sws {
			if sw.IsSame(itemcode) {
				sw.MergeWareHouse(itemcode, area, total)
				flag = true
			}
		}

		if !flag {
			sw := &DivideWarehouse{ItemCode: itemcode, WarehouseInfo: map[string]int{area: total}}
			sws = append(sws, sw)
		}

	}
	return DividItems(sws), nil
}

func (sw *DivideWarehouse) IsSame(code string) bool {
	if sw.ItemCode == code {
		return true
	}
	return false
}

func (sw *DivideWarehouse) MergeWareHouse(code, area string, total int) {
	if val, exit := sw.WarehouseInfo[area]; exit {
		sw.WarehouseInfo[area] = val + total
	} else {
		sw.WarehouseInfo[area] = total
		sw.IsWarehouses = true
	}
}

//把相同商品碼的檔案進行分倉
func (sw *DivideWarehouse) Sub(sf *excelgo.Sourc, items []*model.WarehouseItemInfo) error {

	if len(items) == 0 {
		return nil
	}
	//無多倉
	if !sw.IsWarehouses {
		for _, item := range items {
			for Id := range sw.WarehouseInfo {
				item.WarehouseID = Id
			}
		}
	} else {
		itemcountcol := sf.GetCol(model.Environment.IDS.TotalSpan)
		sw.col = itemcountcol.Col
		//只計算N2倉
		count, _ := sw.WarehouseInfo[Next2]

		index := len(items)

		for {
			if index = sw.sub(count, items[:index]); index != -1 {
				if index == 0 {
					items[index].WarehouseID = MixNext
					break
				}
				//當 Ｎ2倉 無法把商品分配至0品項時 將會出現一項出貨指示有多倉
				if !sw.isEnough(items[:index], count) {
					for _, item := range items[:index] {
						item.WarehouseID = Next2
					}
					items[index].WarehouseID = MixNext
					break
				}

			} else {
				break
			}
		}
		//剩餘填上Ｎ3
		for _, item := range items {
			if item.WarehouseID == "" {
				item.WarehouseID = Next3
			}
		}
	}
	return nil
}

func (sw *DivideWarehouse) sub(count int, items []*model.WarehouseItemInfo) int {
	max, index := sw.find(0, len(items)-1, len(items)/2, count, items)

	if count-max < 0 {
		return index
	} else if count-max > 0 {
		//使用append底層仍然是用同一個數組 需要使用ＣＯＰＹ
		var newitems []*model.WarehouseItemInfo = make([]*model.WarehouseItemInfo, len(items)-1)
		copy(newitems, append(items[:index], items[index+1:]...))

		if flag := sw.sub(count-max, newitems); flag == -1 { //flag正確為0時將會紀錄 item的倉別
			items[index].WarehouseID = Next2
			return flag
		} else {
			return index
		}
	} else { //count-max == 0
		items[index].WarehouseID = Next2
		return -1
	}
}

//返回Target 如果沒有Target返回離Target最接近的小於數,以及座標
func (sw *DivideWarehouse) find(left, right, mid, target int, items []*model.WarehouseItemInfo) (int, int) {
	if items[mid].Data[sw.col].(int) == target {
		return target, mid
	} else if left >= right {
		if items[mid].Data[sw.col].(int) > target && mid-1 >= 0 { //此時mid有可能為0
			return items[mid-1].Data[sw.col].(int), mid - 1
		}
		return items[mid].Data[sw.col].(int), mid
	} else if items[mid].Data[sw.col].(int) > target {
		return sw.find(left, mid-1, (left+mid-1)/2, target, items)
	} else if items[mid].Data[sw.col].(int) < target {
		return sw.find(mid+1, right, (right+mid+1)/2, target, items)
	}
	return -1, -1
}

func (sw *DivideWarehouse) isEnough(items []*model.WarehouseItemInfo, count int) bool {
	sum := 0
	for _, item := range items {
		sum += item.Data[sw.col].(int)
	}
	return count <= sum
}
