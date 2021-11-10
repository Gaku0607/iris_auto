package process

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/slicer"
	"github.com/Gaku0607/iris_auto/store"
)

type SplitFiles struct {
	se model.SpiltAndExportParms
}

func NewSplitFiles() *SplitFiles {
	s := &SplitFiles{}
	s.se = model.Environment.SE
	return s
}

//切割指定內容以及匯出
func (sf *SplitFiles) SplitAndExport(c *augo.Context) {

	sourc, _ := c.Get(model.SOURCE_KEY)

	s := sourc.(*excelgo.Sourc)

	exportfn := func(goods [][]string, path, format string) {

		base := strings.Replace(strings.Split(filepath.Base(path), ".")[0], "【IRIS OHYAMA】", "", 1)
		path = filepath.Join(model.Result_Path, fmt.Sprintf(format, base))
		path = excelgo.CheckFileName(path)

		data, info, err := sf.splitAndexport(s, goods)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		c.Set(filepath.Base(path), info)

		if err = store.Store.Export(path, data); err != nil {
			c.AbortWithError(err)
			return
		}
	}

	for _, path := range c.Request.Files {

		f, err := excelgo.OpenFile(path)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		if err := s.Init(f); err != nil {
			c.AbortWithError(err)
			return
		}

		cf := sf.getCutMethodByPath(path)
		cut := slicer.NewCutFile(s)
		bigs, normals, err := cut.Cut(s.GetRows(), cf)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		//big_goods
		{
			exportfn(bigs, path, sf.se.BigsPathBase)
		}

		//normal_goods
		{
			exportfn(normals, path, sf.se.NormalPathBase)
		}
	}
	return
}

func (sf *SplitFiles) splitAndexport(s *excelgo.Sourc, rows [][]string) ([][]interface{}, string, error) {
	data, err := s.Transform(rows)
	if err != nil {
		return nil, "", err
	}

	s.Sort(data)
	if err = s.Sum(data); err != nil {
		return nil, "", err
	}
	return s.SetHeader(data), fmt.Sprintf("合計数: %d", s.GetCol("合計数").Total), nil
}

//依地址判斷返回需要的切割方式
func (sf *SplitFiles) getCutMethodByPath(Path string) slicer.Slicer {

	var cf slicer.Slicer

	base := path.Base(Path)

	if strings.Contains(base, sf.se.Identify.(string)) {
		cf = slicer.NewMomoCut(
			slicer.NewContentCut(sf.se.SlicerParms.Contains),
			slicer.NewSizeCut(sf.se.SlicerParms.Size),
		)
	} else {
		cf = slicer.NewSizeCut(sf.se.SlicerParms.Size)
	}

	return cf
}
