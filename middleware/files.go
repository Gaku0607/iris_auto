package middleware

import (
	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/tool"
)

func InitFiles(c *augo.Context) {
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
}
