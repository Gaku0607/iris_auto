package process

import (
	"math"
	"strconv"
	"strings"

	"github.com/Gaku0607/excelgo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/model"
)

type DetailQC struct {
	IDS model.ImportDocumentsParms
}

func NewDetailQC() *DetailQC {
	return &DetailQC{IDS: model.Environment.IDS}
}

//獲取穩打的宅配細項
func (d *DetailQC) DividWarehouse(sf, csv *excelgo.Sourc) ([][]interface{}, []*iris.DeliveryDetail, error) {

	itemcodecol := sf.GetCol(d.IDS.ItemCodeSpan)
	totalcol := sf.GetCol(d.IDS.TotalSpan)

	rows, err := sf.Transform(sf.GetRows())
	if err != nil {
		return nil, nil, err
	}
	//依序 合計數.商品碼 進行排序後計算
	excelgo.Sort(rows, totalcol.Col, excelgo.PositiveOrder)
	excelgo.Sort(rows, itemcodecol.Col, excelgo.PositiveOrder)

	sws, err := iris.NewDivideWarehouses(csv)
	if err != nil {
		return nil, nil, err
	}

	var dds []*iris.DeliveryDetail = make([]*iris.DeliveryDetail, len(rows))
	for i := range dds {
		dd := &iris.DeliveryDetail{}
		dds[i] = dd
	}

	return rows, dds, sws.Sub(rows, dds, sf)
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
