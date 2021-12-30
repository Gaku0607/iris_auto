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
	TQC                  model.TripartiteQC
	tripartiteStatusList map[string][][]interface{}
}

func NewZhaipeiQC(detail *DetailQC) *ZhaipeiQC {
	z := &ZhaipeiQC{DetailQC: detail, TQC: model.Environment.TF.TripartiteQC}
	z.tripartiteStatusList = make(map[string][][]interface{}, len(z.TQC.TripartiteStatusList))
	for _, status := range z.TQC.TripartiteStatusList {
		z.tripartiteStatusList[status] = make([][]interface{}, 0)
	}
	return z
}

func (z *ZhaipeiQC) ZhaipeiQC(IsNormalQC bool) augo.HandlerFunc {
	filebase := z.IDS.ZhaipeiMergeBox.MasterFileBase
	return func(c *augo.Context) {

		sourc, _ := c.Get(model.SOURCE_KEY)
		confirmedlist, _ := c.Get(model.CONFIRMED_LIST)
		csvsourc, _ := c.Get(model.CSV_KEY)

		s := sourc.(*excelgo.Sourc)
		z.ConfirmedList = confirmedlist.([]string)

		//當為第三方時
		if !IsNormalQC {
			s.SpanSorts = z.IDS.ZhaipeiMergeBox.ThirdPartySort
			filebase = z.IDS.ZhaipeiMergeBox.ThirdPartyMasterFileBase
		}

		rows, err := z.zhaipeiQC(s, csvsourc.(*excelgo.Sourc))
		if err != nil {
			c.AbortWithError(err)
			return
		}

		//設置ＩＤ
		setNumID(s, rows)

		filename := filepath.Join(model.Result_Path, fmt.Sprintf(filebase, time.Now().Format("0102-15-04")))
		filename = excelgo.CheckFileName(filename)

		if err := store.Store.ExportZhaipeiFiles(filename, s, rows); err != nil {
			c.AbortWithError(err)
		}

	}
}

func (z *ZhaipeiQC) TripartiteQC(c *augo.Context) {
	sourc, _ := c.Get(model.SOURCE_KEY)
	confirmedlist, _ := c.Get(model.CONFIRMED_LIST)
	csvsourc, _ := c.Get(model.CSV_KEY)
	goodstatus, _ := c.Get(model.TRIPARTITE_STATUS_LIST)

	csv := csvsourc.(*excelgo.Sourc)
	s := sourc.(*excelgo.Sourc)
	gs := goodstatus.(map[string]string)
	z.ConfirmedList = confirmedlist.([]string)

	codecol := s.GetCol(z.IDS.UniqueCodeSpan)
	rows, err := z.zhaipeiQC(s, csv)
	if err != nil {
		c.AbortWithError(err)
		return
	}
	//重置三方表單狀態表內容
	defer z.resetTripartiteStatusList()

	for _, row := range rows {
		status, ok := gs[tool.GetUniqueCode(row[codecol.Col].(string))]
		if !ok {
			status = model.TRIPARTITE_STATUS_NULL
		}

		z.tripartiteStatusList[status] = append(z.tripartiteStatusList[status], row)
	}

	t := time.Now().Format("01-02")

	for status, rows := range z.tripartiteStatusList {

		if len(rows) == 0 {
			continue
		}

		//設置ＩＤ
		setNumID(s, rows)

		filename := filepath.Join(model.Result_Path, fmt.Sprintf("%s %s-QC.xlsx", t, status))
		filename = excelgo.CheckFileName(filename)

		if err := store.Store.ExportZhaipeiFiles(filename, s, rows); err != nil {
			c.AbortWithError(err)
			return
		}

	}

}

func (z *ZhaipeiQC) zhaipeiQC(s, csv *excelgo.Sourc) ([][]interface{}, error) {
	rows, dds, err := z.DividWarehouse(s, csv)
	if err != nil {
		return nil, err
	}

	if err = z.setZhaipeiDetail(rows, dds, s); err != nil {
		return nil, err
	}

	if err = z.MergeBox(s, rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (z *ZhaipeiQC) setZhaipeiDetail(rows [][]interface{}, dds []*dividewarehouse.DeliveryDetail, sf *excelgo.Sourc) error {

	uniquecodecol := sf.GetCol(z.IDS.UniqueCodeSpan)
	jancodecol := sf.GetCol(z.IDS.ZhaipeiMergeBox.JanCodeSpan)

	//新增zhaipei_delivery_detail的欄位
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

	return nil
}

//重置三方表單狀態表
func (z *ZhaipeiQC) resetTripartiteStatusList() {
	for status := range z.tripartiteStatusList {
		z.tripartiteStatusList[status] = make([][]interface{}, 0)
	}
}

//設置序號
func setNumID(s *excelgo.Sourc, rows [][]interface{}) {
	s.Sort(rows)
	//設置序號
	for i, row := range rows {
		row[0] = strconv.Itoa(i + 1)
	}
}
