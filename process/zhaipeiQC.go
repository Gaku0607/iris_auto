package process

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/dividewarehouse"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/store"
	"github.com/Gaku0607/iris_auto/tool"
)

type ZhaipeiQC struct {
	*DetailQC
}

func NewZhaipeiQC(detail *DetailQC) *ZhaipeiQC {
	z := &ZhaipeiQC{DetailQC: detail}
	return z
}

//宅配通ＱＣ
func (z *ZhaipeiQC) OrginZhaipeiQC(c *augo.Context) {
	sourc, _ := c.Get(model.SOURCE_KEY)
	csvsourc, _ := c.Get(model.CSV_KEY)
	confirmedlist, _ := c.Get(model.CONFIRMED_LIST)

	z.ConfirmedList = confirmedlist.([]string)
	if err := z.exportQC(sourc.(*excelgo.Sourc), csvsourc.(*excelgo.Sourc), z.IDS.ZhaipeiMergeBox.MasterFileBase); err != nil {
		c.AbortWithError(err)
	}
}

//第三方宅配ＱＣ
func (z *ZhaipeiQC) ThirdPartyQC(c *augo.Context) {
	sourc, _ := c.Get(model.SOURCE_KEY)
	confirmedlist, _ := c.Get(model.CONFIRMED_LIST)
	csvsourc, _ := c.Get(model.CSV_KEY)

	s := sourc.(*excelgo.Sourc)
	z.ConfirmedList = confirmedlist.([]string)

	s.SpanSorts = z.IDS.ZhaipeiMergeBox.ThirdPartySort

	if err := z.exportQC(s, csvsourc.(*excelgo.Sourc), z.IDS.ZhaipeiMergeBox.ThirdPartyMasterFileBase); err != nil {
		c.AbortWithError(err)
	}
}

func (z *ZhaipeiQC) exportQC(s, csv *excelgo.Sourc, filebase string) error {

	rows, dds, err := z.DividWarehouse(s, csv)
	if err != nil {
		return err
	}

	if err = z.setZhaipeiDetail(rows, dds, s); err != nil {
		return err
	}

	if err = z.MergeBox(s, rows); err != nil {
		return err
	}

	filename := filepath.Join(model.Result_Path, fmt.Sprintf(filebase, time.Now().Format("0102-15-04")))
	filename = excelgo.CheckFileName(filename)

	if err = store.Store.ExportZhaipeiFiles(filename, s, rows); err != nil {
		return err
	}
	return nil
}

func (z *ZhaipeiQC) setZhaipeiDetail(rows [][]interface{}, dds []*dividewarehouse.DeliveryDetail, sf *excelgo.Sourc) error {

	uniquecodecol := sf.GetCol(z.IDS.UniqueCodeSpan)
	jancodecol := sf.GetCol(z.IDS.ZhaipeiMergeBox.JanCodeSpan)
	//新增zhaipei_delivery_detail的欄位
	{
		for i, row := range rows {
			//取注文番號前14碼
			dds[i].DeliveryOrder = tool.GetUniqueCode(row[uniquecodecol.Col].(string))
			//取商品尾碼
			icode := row[jancodecol.Col].(string)
			dds[i].ItemCode = icode[len(icode)-1:]

			rows[i] = append(dds[i].GetZhaipeiRow(), row...)
		}

		sf.Cols = append(sf.Cols, z.IDS.ZhaipeiMergeBox.NewCols...)

		if err := sf.ResetCol(append(z.IDS.ZhaipeiMergeBox.NewHeaders, sf.Headers...)); err != nil {
			return err
		}
	}

	sf.Sort(rows)
	//設置序號
	for i, row := range rows {
		row[0] = strconv.Itoa(i + 1)
	}

	return nil
}
