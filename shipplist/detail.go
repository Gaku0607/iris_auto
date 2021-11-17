package shipplist

import (
	"strconv"
)

type ShippDetailes []*ShippingDetail

func (s ShippDetailes) IsRecipientSame(recipient string) (*ShippingDetail, bool) {
	for _, d := range s {
		if d.Recipient == recipient {
			return d, true
		}
	}
	return nil, false
}

func (s ShippDetailes) Merge() error {
	for _, d := range s {
		if err := d.merge(); err != nil {
			return nil
		}
	}
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

//返回日期
func (s *ShippingDetail) Date() string {
	return s.date
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
				i--
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
