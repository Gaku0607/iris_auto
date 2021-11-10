package process

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/store"
)

type WendaQC struct {
	*DetailQC
}

func NewWendaQC(deliverydetail *DetailQC) *WendaQC {
	return &WendaQC{
		DetailQC: deliverydetail,
	}
}

func (wc *WendaQC) WendaMergeBoxAndExportList(c *augo.Context) {

	sourc, _ := c.Get(model.SOURCE_KEY)
	csvsourc, _ := c.Get(model.CSV_KEY)

	s := sourc.(*excelgo.Sourc)
	csv := csvsourc.(*excelgo.Sourc)

	rows, dds, err := wc.DividWarehouse(s, csv)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	//設置穩達宅配細節
	if err := wc.setWendaDetail(rows, dds, s); err != nil {
		c.AbortWithError(err)
		return
	}

	// return nil
	if err = wc.MergeBox(s, rows); err != nil {
		c.AbortWithError(err)
		return
	}

	//設置“多件”
	wc.setCoutnStr(rows, s)

	filename := filepath.Join(model.Result_Path, fmt.Sprintf(wc.IDS.WendaMergeBox.MasterFileBase, time.Now().Format("0102-15-04")))
	filename = excelgo.CheckFileName(filename)

	if err := store.Store.ExportWendaFiles(wc.IDS.WendaMergeBox.WendaFormatPath, filename, s, rows); err != nil {
		c.AbortWithError(err)
	}
}

func (wc *WendaQC) setWendaDetail(rows [][]interface{}, dds []*iris.DeliveryDetail, sf *excelgo.Sourc) error {

	// 郵遞區號欄位
	postalcol := sf.GetCol(wc.IDS.PostalCodeSpan)

	//新增wenda_delivery_detail的欄位
	for i, row := range rows {

		postalcode := strings.Trim(row[postalcol.Col].(string), " ")
		if len(postalcode) < 3 {
			return errors.New("postalcode is not format")
		}

		//穩達的宅配區號只取前3碼
		postalcode = postalcode[:3]
		row[postalcol.Col] = postalcode

		if val, exit := model.DeliveryArea[postalcode]; exit {
			dds[i].AriaCode = val.AreaCode
			dds[i].Place = val.Place
		} else {
			return fmt.Errorf("Input %s ,DeliveryCode is not exit", row[postalcol.Col])
		}
		rows[i] = append(dds[i].GetWendaRow(), row...)
	}

	sf.Cols = append(sf.Cols, wc.IDS.WendaMergeBox.NewCols...)

	if err := sf.ResetCol(append(wc.IDS.WendaMergeBox.NewHeaders, sf.Headers...)); err != nil {
		return err
	}

	//進行匯出時的排序
	sf.Sort(rows)

	//設置序號以及客代

	//時間搓
	t := time.Now().Format("0601021504-")
	for i, row := range rows {
		ordernum := ""
		orderbase := strconv.Itoa(i + 1)
		for i := 0; i < 3-len(orderbase); i++ {
			ordernum += "0"
		}
		ordernum += orderbase
		row[0], row[1] = orderbase, wc.IDS.WendaMergeBox.CustomerCode+t+ordernum
	}

	return nil
}

//設置多箱欄位內容
func (wc *WendaQC) setCoutnStr(rows [][]interface{}, sf *excelgo.Sourc) {

	boxcountcol := sf.GetCol(wc.IDS.BoxCountSpan)
	countstr := sf.GetCol(wc.IDS.WendaMergeBox.CountStrSpan)

	for i := range rows {
		if rows[i][boxcountcol.Col].(float64) > 1 {
			rows[i][countstr.Col] = "多件"
		}
	}
}
