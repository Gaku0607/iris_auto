package process

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/shipplist"
	"github.com/Gaku0607/iris_auto/store"
	"github.com/xuri/excelize/v2"
)

type ShippList struct {
	sl model.ShippingListParms
}

func NewShippList() *ShippList {
	sl := &ShippList{}
	sl.sl = model.Environment.SL
	return sl
}

//製作出倉單
func (sl *ShippList) ExportShippList(c *augo.Context) {

	sourc, _ := c.Get(model.SOURCE_KEY)
	csvsourc, _ := c.Get(model.CSV_KEY)
	s := sourc.(*excelgo.Sourc)
	csv := csvsourc.(*excelgo.Sourc)

	if len(s.Rows) <= 0 {
		c.AbortWithError(errors.New("data is nil"))
		return
	}

	t := s.Rows[0][sl.sl.DateIndex]
	loc, _ := time.LoadLocation("Local")
	d, _ := time.ParseInLocation("20060102", t, loc)
	date := d.Format("20060102")

	num, err := sl.getOrderNumByDate(date)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	shipp, _, err := sl.shipplist(csv, s, date, num)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	var tcol []string
	for _, col := range csv.Cols {
		i := ""
		if len(col.TCol) > 0 {
			i = col.TCol[0].TColStr
		}
		tcol = append(tcol, i)
	}

	for _, detail := range shipp {

		f, err := excelize.OpenFile(sl.sl.FileFormatPath)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		path := filepath.Join(model.Result_Path, fmt.Sprintf(sl.sl.OutputFileFormat, date, detail.Recipient))

		path = excelgo.CheckFileName(path)

		if err = store.Store.ExportShippList(f, detail, tcol, path); err != nil {
			c.AbortWithError(err)
			return
		}

	}
	return
}

func (sl *ShippList) shipplist(csv, s *excelgo.Sourc, date string, ordernum int) (shipplist.ShippDetailes, int, error) {

	csvsourc := csv
	commoditysourc := s

	comm_codecol := commoditysourc.GetCol(sl.sl.CommoditySourc.CodeSpan).Col
	comm_jancodecol := commoditysourc.GetCol(sl.sl.CommoditySourc.JANCodeSpan).Col

	csv_codecol := csvsourc.GetCol(sl.sl.CsvSourc.CodeSpan).Col
	csv_addcol := csvsourc.GetCol(sl.sl.CsvSourc.AddrSpan).Col
	csv_areacol := csvsourc.GetCol(sl.sl.CsvSourc.AreaSpan).Col
	csv_telcol := csvsourc.GetCol(sl.sl.CsvSourc.TelSpan).Col
	csv_recipiencol := csvsourc.GetCol(sl.sl.CsvSourc.RecipientSpna).Col

	var JanCodecol int
	shippfile := shipplist.ShippDetailes{}
	var row []interface{}

	for i, col := range csvsourc.Cols {
		if csv_codecol == col.Col {
			JanCodecol = i
		}
	}

	for _, r := range csvsourc.GetRows() {

		row = make([]interface{}, len(csvsourc.Cols))

		//把商品馬替換成13碼
		for _, cr := range commoditysourc.GetRows() {
			if r[csv_codecol] == cr[comm_codecol] {
				r[csv_codecol] = cr[comm_jancodecol]
				break
			}
		}

		for i, col := range csvsourc.Cols {
			val, err := col.TransferFormat(r[col.Col])
			if err != nil {
				return nil, 0, err
			}
			row[i] = val
		}

		recipient := r[csv_recipiencol]

		area := iris.GetAria(r[csv_areacol])

		if detail, exit := shippfile.IsRecipientSame(recipient); exit {
			if Area, exit := detail.IsAreaSame(area); exit {
				Area.Add(row)
			} else {
				detail.AddArea(area, row)
			}
		} else {
			shipp := shipplist.NewShippingDetail(
				recipient,
				area,
				date,
				row,
				ordernum,
				JanCodecol,
				sl.sl.TotalIndex,
				sl.sl.PackingMethodIndex,
			)
			shipp.Addr = r[csv_addcol]
			if len(r[csv_telcol]) != 0 && string(r[csv_telcol][0]) != "0" {
				r[csv_telcol] = "0" + r[csv_telcol]
			}
			shipp.Tel = r[csv_telcol]
			shippfile = append(shippfile, shipp)
			ordernum++
		}
	}

	return shippfile, ordernum, shippfile.Merge()
}

func (sl *ShippList) getOrderNumByDate(date string) (int, error) {
	return 0, nil
}
