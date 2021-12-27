package middleware

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/tool"
)

func InitTripartiteForm() augo.HandlerFunc {
	datecol := 0
	statuscol := 0
	uniquecodecol := 0
	newdate := ""
	sourc := model.Environment.TF.TripartiteQC.Sourc
	removeAbnormalRowfn := removeAbnormalRow(sourc.SheetName)
	return func(c *augo.Context) {
		gs := make(map[string]string)
		for _, file := range c.Request.Files {

			if !tool.IsTripartiteQCMethod(file) {
				continue
			}

			f, err := excelgo.OpenFile(file)
			if err != nil {
				c.AbortWithError(err)
				return
			}

			if err := sourc.Init(f); err != nil {
				c.AbortWithError(err)
				return
			}

			rows := sourc.Rows
			if len(rows) == 0 {
				continue
			}

			datecol = sourc.GetCol(model.Environment.TF.TripartiteQC.DateSpan).Col
			statuscol = sourc.GetCol(model.Environment.TF.TripartiteQC.StatusSpan).Col
			uniquecodecol = sourc.GetCol(model.Environment.TF.TripartiteQC.UniqueCodeSpan).Col

			//刪除底部所有空白方便計算
			rows = removeBlankRow(rows, uniquecodecol)

			//刪除異常行方便計算
			rows, err = removeAbnormalRowfn(f, rows, sourc.GetStartBlankRowsCount())
			if err != nil {
				c.AbortWithError(err)
				return
			}

			//刪除不必要的信息後 最後為最新日期
			if newdate == "" {
				newdate = rows[len(rows)-1][datecol]
			} else {
				t, err := parseDate(newdate)
				if err != nil {
					c.AbortWithError(err)
					return
				}

				date, err := parseDate(rows[len(rows)-1][datecol])
				if err != nil {
					c.AbortWithError(err)
					return
				}

				if !t.Equal(date) {
					c.AbortWithError(errors.New("The latest date is inconsistent"))
					return
				}
			}

			for i := len(rows) - 1; i > 0; i-- {
				row := rows[i]
				if newdate != row[datecol] {
					break
				}

				if model.IsTripartiteStatus(row[statuscol]) {
					gs[row[uniquecodecol]] = row[statuscol]
				}
			}
		}

		c.Set(model.TRIPARTITE_STATUS_LIST, gs)
		c.SetLogKey("Latest date", newdate)
		sourc.Rows = nil
	}
}

func parseDate(datestr string) (time.Time, error) {
	datestr = strings.TrimSpace(datestr)
	format := ""
	if b, _ := regexp.MatchString("^[0-9]{1,2}/[0-9]{1,2}$", datestr); b {
		format = "01/02"
	} else {
		if b, _ = regexp.MatchString("^[0-9]{1,2}月[0-9]{1,2}日$", datestr); b {
			format = "01月02日"
		}
	}

	if format == "" {
		return time.Time{}, fmt.Errorf(`Unable to parse the date: "%s"`, datestr)
	}

	return time.Parse(format, datestr)
}

//刪除Sheet底部所有空白行
func removeBlankRow(rows [][]string, uniquecol int) [][]string {
	index := 0
	for i := len(rows) - 1; i > 0; i-- {
		if rows[i][uniquecol] != "" {
			index = i
			break
		}
	}
	return rows[:index+1]
}

//刪除異常行 異常行會已Color做標記 當Color為“”時為正常
//blankcount為起始的以被刪除的空白行數
func removeAbnormalRow(sheetname string) func(f excelgo.FormFile, rows [][]string, blankcount int) ([][]string, error) {

	return func(f excelgo.FormFile, rows [][]string, blankcount int) ([][]string, error) {
		//+1補上被除掉的Header
		blankcount++

		index := 0
		xlsxf, ok := f.(*excelgo.XlsxFile)
		if !ok {
			return nil, errors.New("removeAbnormalRow is failed")
		}
		for i := len(rows) + blankcount; i > 0; i-- {
			color := tool.GetCellBgColor(xlsxf.File, sheetname, "B"+strconv.Itoa(i))
			if color == "" || color == "FFFFFF" {
				index = i - blankcount
				break
			}
		}
		return rows[:index], nil
	}
}

func InitQCSourceFiles(c *augo.Context) {
	sourc, _ := c.Get(model.SOURCE_KEY)
	s := sourc.(*excelgo.Sourc)

	var (
		confirmedlist      []string
		totalrows          [][]string
		ToBeConfirmedSheet = model.Environment.IDS.ToBeConfirmedSheet
	)

	files, err := excelgo.OpenFiles(c.Request.Files)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	for _, f := range files {
		if tool.IsTripartiteQCMethod(f.Path()) {
			continue
		}

		if tool.IsCsvFormat(f.Path()) {

			if xlsxcsv, ok := f.(*excelgo.XlsxFile); ok && xlsxcsv.File.GetSheetIndex(ToBeConfirmedSheet) != -1 {
				list, _ := xlsxcsv.File.GetRows(ToBeConfirmedSheet)
				confirmedlist = make([]string, len(list))
				for i, code := range list {
					if len(code) > 0 {
						confirmedlist[i] = code[0]
					}
				}
			}

			continue
		}

		if err := s.Init(f); err != nil {
			c.AbortWithError(err)
			return
		}

		totalrows = append(totalrows, s.Rows...)
	}

	s.Rows = totalrows
	c.Set(model.SOURCE_KEY, s)
	c.Set(model.CONFIRMED_LIST, confirmedlist)
	c.SetLogKey(model.CONFIRMED_LIST, confirmedlist)
}
