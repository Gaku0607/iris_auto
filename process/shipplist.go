package process

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/shipplist"
	"github.com/Gaku0607/iris_auto/store"
	"github.com/Gaku0607/iris_auto/tool"
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

	oldnum, err := sl.getOrderNumByDate(date)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	shipp, num, err := sl.shipplist(csv, s, date, oldnum)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	if err := sl.saveOrderNumByDate(date, num); err != nil {
		c.AbortWithError(err)
		return
	}

	c.Set("order_num", fmt.Sprintf("%s = %d --> %d", date, oldnum, num))

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
			ordernum++
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
		}
	}

	return shippfile, ordernum, shippfile.Merge()
}

func (sl *ShippList) getOrderNumByDate(targetdate string) (int, error) {

	if targetdate == "00010101" {
		return 0, nil
	}

	file, err := os.Open(sl.sl.HistoryEnvPath)
	defer file.Close()
	if err != nil {
		return 0, err
	}

	bufReader := bufio.NewReader(file)

	for line, _, err := bufReader.ReadLine(); err != io.EOF; line, _, err = bufReader.ReadLine() {

		if len(line) == 0 {
			continue
		}

		if tool.IsAnnotaion(string(line)) {
			continue
		}

		linestr := strings.SplitN(string(line), "=", 2)
		if len(linestr) != 2 {
			continue
		}

		date := linestr[0]
		count := linestr[1]
		if date == targetdate {
			return strconv.Atoi(count)
		}

	}

	return 0, nil
}

func (sl *ShippList) saveOrderNumByDate(targetdate string, count int) error {

	if targetdate == "00010101" {
		return nil
	}

	file, err := os.OpenFile(sl.sl.HistoryEnvPath, os.O_RDWR, 0777)
	defer file.Close()
	if err != nil {
		return err
	}

	olddata, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	data := []byte(targetdate + "=" + strconv.Itoa(count) + augo.GetNewLine())

	if len(olddata) == 0 {
		_, err := file.Write(data)
		return err
	}

	rows := bytes.Split(olddata, []byte(augo.GetNewLine()))
	idx := 0
	rowsidx := 0
	linelen := len([]byte(augo.GetNewLine()))

	for i, row := range rows {

		rowsidx = i + 1
		idx += len(row) + linelen

		if tool.IsAnnotaion(string(row)) {
			continue
		}

		linestr := bytes.SplitN(row, []byte("="), 2)
		if len(linestr) != 2 {
			continue
		}

		if string(linestr[0]) == targetdate {

			idx = idx - (len(row) + linelen)
			rowsidx--

			if len(linestr[1]) != len([]byte(strconv.Itoa(count))) {
				//......
				for _, row := range rows[i+1:] {
					data = append(data, row...)
					data = append(data, []byte(augo.GetNewLine())...)
				}
			}
			break
		}

	}

	//當迴圈全跑完時會多一個換行服 需要刪除
	if rowsidx == len(rows) {
		idx--
	}

	//確認前一個byte是否為換行符
	if string(olddata[idx-1]) != augo.GetNewLine() {
		data = append([]byte(augo.GetNewLine()), data...)
	}

	if _, err = file.WriteAt(data, int64(idx)); err != nil {
		return err
	}

	return nil
}
