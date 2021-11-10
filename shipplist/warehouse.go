package shipplist

import (
	"errors"
	"fmt"
	"math"

	"github.com/Gaku0607/iris_auto/model"
)

type Areas map[string]*Warehouse

//確認是否有地區相同
func (a Areas) IsAreaSame(area string) (*Warehouse, bool) {
	w, ok := a[area]
	return w, ok
}

//新增倉庫
func (a Areas) addArea(area string, data []interface{}) error {

	if _, ok := a.IsAreaSame(area); ok {
		return errors.New("Cannot add the sane area")
	}

	wh := &Warehouse{}
	wh.Area = area
	wh.Rows = append(wh.Rows, data)
	a[area] = wh
	return nil
}

//新增商品至指定倉庫
func (a Areas) addItem(area string, item []interface{}) error {

	wh, ok := a.IsAreaSame(area)
	if !ok {
		return fmt.Errorf("input:%s area is not exist", area)
	}
	wh.add(item)
	return nil
}

//該倉庫的所有資訊
type Warehouse struct {
	*ShippingDetail
	Area string
	Rows [][]interface{}
}

func (w *Warehouse) Total(rows [][]interface{}) float64 {
	var total float64
	for _, row := range rows {
		total += row[w.totalcol].(float64)
	}
	return total
}

func (w *Warehouse) Format() error {
	for _, row := range w.Rows {
		quantity := row[w.quantitycol]
		total := row[w.totalcol]
		row[w.quantitycol] = model.Environment.SL.PackingMethod
		row[w.totalcol] = math.Ceil(float64(total.(int)) / float64(quantity.(int)))
	}
	return nil
}

//刪除指定位置元素
func (w *Warehouse) deleteItem(index int) {
	w.Rows = append(w.Rows[:index], w.Rows[index+1:]...)
}

//添加元素至同倉庫 如果JANCODE相同的商品已存在則增加總量
func (w *Warehouse) Add(data []interface{}) {
	var flag bool
	for _, r := range w.Rows {
		if r[w.janCodecol] == data[w.janCodecol] {
			r[w.totalcol] = r[w.totalcol].(int) + data[w.totalcol].(int)
			flag = true
		}
	}
	if !flag {
		w.add(data)
	}
}

//直接添加元素至該倉
func (w *Warehouse) add(data []interface{}) {
	w.Rows = append(w.Rows, data)
}
