package process

import (
	"math"
	"strconv"
	"strings"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/dividewarehouse"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/tool"
)

type DetailQC struct {
	IDS           model.ImportDocumentsParms
	ConfirmedList []string
}

func NewDetailQC() *DetailQC {
	return &DetailQC{IDS: model.Environment.IDS}
}

//獲取穩打的宅配細項
func (d *DetailQC) DividWarehouse(sf, csv *excelgo.Sourc) ([][]interface{}, []*dividewarehouse.DeliveryDetail, error) {

	itemcodecol := sf.GetCol(d.IDS.ItemCodeSpan)
	totalcol := sf.GetCol(d.IDS.TotalSpan)
	uniquecodecol := sf.GetCol(d.IDS.UniqueCodeSpan)

	rows, err := sf.Transform(sf.Rows)
	if err != nil {
		return nil, nil, err
	}

	//獲取ＱＣ單相關信息欄位
	dds, err := d.getDetail(len(rows))
	if err != nil {
		return nil, nil, err
	}

	sws, err := dividewarehouse.NewDivideWarehouses(csv)
	if err != nil {
		return nil, nil, err
	}

	var orders dividewarehouse.MultipleItemOrders

	for i, row := range rows {

		if d.isConfirmedList(row[uniquecodecol.Col].(string)) {
			continue
		}

		itemcode := row[itemcodecol.Col].(string)
		ordercode := tool.GetUniqueCode(row[uniquecodecol.Col].(string))
		count := row[totalcol.Col].(int)

		if itemorder, b := orders.IsItemExist(itemcode); b {
			itemorder.AddOrderForm(ordercode, count, i)
		} else {
			orders = append(orders, dividewarehouse.NewItemOrder(itemcode, ordercode, count, i))
		}

	}

	return rows, dds, sws.Divide(orders, dds, sf)
}

//確認是否為待確認名單
func (d *DetailQC) isConfirmedList(code string) bool {
	for _, confirmedcode := range d.ConfirmedList {
		if confirmedcode == code {
			return true
		}
	}
	return false
}

//獲取Detail
func (d *DetailQC) getDetail(rowslen int) ([]*dividewarehouse.DeliveryDetail, error) {
	var dds []*dividewarehouse.DeliveryDetail = make([]*dividewarehouse.DeliveryDetail, rowslen)
	for i := range dds {
		dd := &dividewarehouse.DeliveryDetail{}
		dds[i] = dd
	}

	return dds, nil
}

//併箱
func (d *DetailQC) MergeBox(sf *excelgo.Sourc, newrows [][]interface{}) error {

	totalcol := sf.GetCol(d.IDS.TotalSpan)
	quantitycol := sf.GetCol(d.IDS.QuantitySpan)
	remarkcol := sf.GetCol(d.IDS.RemarkSpan)
	uniquecodecol := sf.GetCol(d.IDS.UniqueCodeSpan)
	boxcountcol := sf.GetCol(d.IDS.BoxCountSpan)

	type T struct {
		index    int
		code     string
		total    int
		quantity int
	}

	var (
		tm   map[string][]T = make(map[string][]T)
		data [][]interface{}
	)

	for i, row := range newrows {

		boxcount := math.Ceil(float64(row[totalcol.Col].(int)) / float64(row[quantitycol.Col].(int)))

		if row[quantitycol.Col] != row[totalcol.Col] && row[quantitycol.Col] != "1" {
			if strings.Contains(row[remarkcol.Col].(string), "併箱") || strings.LastIndex(row[uniquecodecol.Col].(string), "A") == len(row[uniquecodecol.Col].(string)) ||
				boxcount > 1 {
				code := strings.TrimRight(row[uniquecodecol.Col].(string), "A")
				tm[code] = append(tm[code], T{index: i, code: row[uniquecodecol.Col].(string), total: row[totalcol.Col].(int), quantity: row[quantitycol.Col].(int)})
				continue
			}
		}
		newrows[i][boxcountcol.Col] = boxcount
	}

	//驗證 併箱
	for _, T := range tm {
		var (
			w     float64 //權重
			flag  bool
			count float64
		)
		for i, t := range T {
			for _, val := range data {
				if t.code == val[0] {
					weights, err := strconv.Atoi(val[3].(string))
					if err != nil {
						return err
					}
					w += float64(weights * t.total)
					flag = true
					break
				}
			}
			if flag && i != len(T) {
				newrows[t.index][boxcountcol.Col] = 0
			} else if !flag {
				count = math.Ceil(float64(t.total) / float64(t.quantity)) //不符合並箱條件
				newrows[t.index][boxcountcol.Col] = count
			} else {
				// count = math.Ceil(w / p.IDS.MergeBoxSourc.Weights)
				newrows[t.index][boxcountcol.Col] = count
			}
		}
	}
	return nil
}
