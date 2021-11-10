package middleware

import (
	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/tool"
)

func InitSourcFiles(c *augo.Context) {
	sourc, _ := c.Get(model.SOURCE_KEY)
	s := sourc.(*excelgo.Sourc)

	var totalrows [][]string

	files, err := excelgo.OpenFiles(c.Request.Files)
	if err != nil {
		c.AbortWithError(err)
		return
	}

	for _, f := range files {

		if tool.IsCsvFormat(f.Path()) {
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
}
