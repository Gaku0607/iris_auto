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

func InitTripartiteQCForm() augo.HandlerFunc {
	datecol := 0
	statuscol := 0
	uniquecodecol := 0
	newdate := ""
	sourc := model.Environment.TF.TripartiteQC.Sourc
	removeAbnormalRowfn := removeAbnormalRow(sourc.SheetName)
	parseDatefn := parseDate()
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
			statuscol = sourc.GetCol(model.Environment.TF.TripartiteQC.GoodsStatusSpan).Col
			uniquecodecol = sourc.GetCol(model.Environment.TF.UniqueCodeSpan).Col

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
				t, err := parseDatefn(newdate)
				if err != nil {
					c.AbortWithError(err)
					return
				}

				newdate = rows[len(rows)-1][datecol]

				date, err := parseDatefn(newdate)
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
					code := strings.TrimSpace(row[uniquecodecol])
					code = tool.GetUniqueCode(code)
					gs[code] = row[statuscol]
				}
			}
		}

		c.Set(model.TRIPARTITE_STATUS_LIST, gs)
		c.SetLogKey("Latest date", newdate)
		sourc.Rows = nil
	}
}

func parseDate() func(string) (time.Time, error) {
	regexpmap := map[string]string{
		"01/02":  "^[0-9]{1,2}/[0-9]{1,2}$",
		"01月02日": "^[0-9]{1,2}月[0-9]{1,2}日$",
		"01-02":  "^[0-9]{1,2}-[0-9]{1,2}$",
	}
	return func(datestr string) (time.Time, error) {
		datestr = strings.TrimSpace(datestr)
		format := ""

		//格式為 01-02-21時  不計算year
		if strings.Count(datestr, "-") > 1 {
			sli := strings.Split(datestr, "-")[:2]
			datestr = strings.Join(sli, "-")
		}

		for dateformat, matchstr := range regexpmap {
			if b, _ := regexp.MatchString(matchstr, datestr); b {
				format = dateformat
				break
			}
		}

		if format == "" {
			return time.Time{}, fmt.Errorf(`Unable to parse the date: "%s"`, datestr)
		}
		return time.Parse(format, datestr)
	}
}

//刪除Sheet底部所有空白行
func removeBlankRow(rows [][]string, uniquecol int) [][]string {
	index := 0
	for i := len(rows) - 1; i > 0; i-- {
		if len(rows[i]) > uniquecol && rows[i][uniquecol] != "" {
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
			color := tool.GetCellBgColor(xlsxf.File, sheetname, "C"+strconv.Itoa(i))
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
