package iris

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Gaku0607/iris_auto/model"
)

//出倉單所需的變量
type WarehouseReceiptDetail struct {
	Party string
	Num   int
	count int
	Date  time.Time
}

func NewWarehouseReceiptDetail(name string, num int) *WarehouseReceiptDetail {
	w := &WarehouseReceiptDetail{}
	w.Party = name
	w.Num = num
	return w
}

func (w *WarehouseReceiptDetail) GetDate() string {
	return w.Date.Format("2006/01/02")
}

type ShippingFile struct {
	Tcol   []int //出倉單資訊所對應的欄位
	Detail []*ShippingDetail
}

func (s *ShippingFile) IsRecipientSame(recipient string) (*ShippingDetail, bool) {
	for _, d := range s.Detail {
		if d.Recipient == recipient {
			return d, true
		}
	}
	return nil, false
}

func (s *ShippingFile) Merge() error {
	for _, d := range s.Detail {
		if err := d.merge(); err != nil {
			return nil
		}
	}
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

//出倉單所有細項
type ShippingDetail struct {
	Recipient string //收件人
	Areas            //出倉商品位置以及信息
	Addr      string //地址
	Tel       string //電話

	ordernum int
	date     string

	janCodecol  int
	totalcol    int
	quantitycol int
}

func NewShippingDetail(recipient, area, date string, data []interface{}, ordernum, JanCodecol, totalcol, quantitycol int) *ShippingDetail {

	s := &ShippingDetail{}
	s.Recipient = recipient
	wh := &Warehouse{}
	wh.ShippingDetail = s
	wh.Area = area
	wh.Rows = append(wh.Rows, data)

	if s.Areas == nil {
		s.Areas = make(Areas)
	}

	s.Areas[area] = wh
	s.ordernum = ordernum
	s.date = date

	s.janCodecol = JanCodecol
	s.totalcol = totalcol
	s.quantitycol = quantitycol

	return s
}

//增添地區
func (s *ShippingDetail) AddArea(area string, data []interface{}) error {
	if err := s.addArea(area, data); err != nil {
		return err
	}
	s.Areas[area].ShippingDetail = s
	return nil
}

//返回單號
func (s *ShippingDetail) OrderNumber() string {
	ordernum := s.date
	orderbase := strconv.Itoa(s.ordernum)

	for i := 0; i < 3-len(orderbase); i++ {
		ordernum += "0"
	}
	return ordernum + orderbase
}

//同品項進行分區合併
func (s *ShippingDetail) merge() error {

	type T struct {
		area      string
		index     int
		data      []interface{}
		remainder int //餘數 (不滿箱時)
	}

	var mergemap map[string][]*T = make(map[string][]*T)

	for _, wh := range s.Areas {
		for i := 0; i < len(wh.Rows); i++ {
			//當總數 餘 入數 不能整除時 需進行合併操作
			remainder := wh.Rows[i][s.totalcol].(int) % wh.Rows[i][s.quantitycol].(int)
			if remainder != 0 {
				mergemap[wh.Rows[i][s.janCodecol].(string)] = append(mergemap[wh.Rows[i][s.janCodecol].(string)],
					&T{area: wh.Area, data: wh.Rows[i], remainder: remainder})
				wh.deleteItem(i)
			}
		}
	}

	for _, t := range mergemap {

		var (
			max   int
			index int
			total int
		)

		//該商品只存在1個 則不用併箱
		if len(t) == 1 {
			s.Areas.addItem(t[index].area, t[index].data)
			continue
		}

		for i, val := range t {
			if max <= val.remainder {
				max = val.remainder
				index = i
			}
			total += val.remainder
		}

		t[index].data[s.totalcol] = t[index].data[s.totalcol].(int) + total

		for _, item := range t {
			r := item.data[s.totalcol].(int) - item.remainder
			if r > 0 {
				item.data[s.totalcol] = r
				s.addItem(item.area, item.data)
			}
		}

	}
	return nil
}
