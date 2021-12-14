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

type SpliteFilesFunc func(c *augo.Context, s *excelgo.Sourc)

//三方分單
func TripartiteSplitFiles(c *augo.Context, s *excelgo.Sourc) {

	var biggoods [][]string
	var normalgoods [][]string
	const bigbase = "三方專車"
	const normalbase = "三方宅配"

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

		cf := getCutMethodByPath(path)
		cut := slicer.NewCutFile(s)
		bigs, normals, err := cut.Cut(s.GetRows(), cf)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		biggoods = append(biggoods, bigs...)
		normalgoods = append(normalgoods, normals...)

	}

	//專車
	msg, err := export(s, biggoods, bigbase+".xlsx")
	if err != nil {
		c.AbortWithError(err)
		return
	}
	c.Set(bigbase, msg)

	//宅配
	msg, err = export(s, normalgoods, normalbase+".xlsx")
	if err != nil {
		c.AbortWithError(err)
		return
	}
	c.Set(normalbase, msg)
}

//基本分單
func OriginSpliteFiles(c *augo.Context, s *excelgo.Sourc) {

	var bigbase string
	var normalbase string

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

		cf := getCutMethodByPath(path)
		cut := slicer.NewCutFile(s)
		bigs, normals, err := cut.Cut(s.GetRows(), cf)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		//專車
		bigbase = getBigGoodBase(path)
		msg, err := export(s, bigs, bigbase)
		if err != nil {
			c.AbortWithError(err)
			return
		}
		c.Set(bigbase, msg)

		//宅配
		normalbase = getNormalGoodBase(path)
		msg, err = export(s, normals, normalbase)
		if err != nil {
			c.AbortWithError(err)
			return
		}
		c.Set(normalbase, msg)

	}
}

type SplitFiles struct {
	se model.SpiltAndExportParms
}

func NewSplitFiles() *SplitFiles {
	s := &SplitFiles{}
	s.se = model.Environment.SE
	return s
}

func (sf *SplitFiles) SplitAndExport(fn SpliteFilesFunc) augo.HandlerFunc {
	return func(c *augo.Context) {
		sourc, _ := c.Get(model.SOURCE_KEY)
		s := sourc.(*excelgo.Sourc)

		fn(c, s)
	}
}

//切割指定內容以及匯出
func export(s *excelgo.Sourc, goods [][]string, base string) (string, error) {

	if len(goods) == 0 {
		return "", nil
	}

	path := filepath.Join(model.Result_Path, base)
	path = excelgo.CheckFileName(path)

	data, err := s.Transform(goods)
	if err != nil {
		return "", err
	}

	msg, err := getSumMsg(s, data)
	if err != nil {
		return "", err
	}

	s.Sort(data)
	data = s.SetHeader(data)
	if err = store.Store.Export(path, data); err != nil {
		return "", err
	}
	return msg, nil
}

//獲取合計數的信息
func getSumMsg(s *excelgo.Sourc, data [][]interface{}) (string, error) {
	if err := s.Sum(data); err != nil {
		return "", err
	}
	return fmt.Sprintf("合計数: %d", s.GetCol("合計数").Total), nil
}

//依地址判斷返回需要的切割方式
func getCutMethodByPath(Path string) slicer.Slicer {

	var cf slicer.Slicer
	var se model.SlicerParms = model.Environment.SE.SlicerParms

	base := path.Base(Path)

	if strings.Contains(base, se.Identify.(string)) {
		cf = slicer.NewMomoCut(
			slicer.NewContentCut(se.Contains),
			slicer.NewSizeCut(se.Size),
		)
	} else {
		cf = slicer.NewSizeCut(se.Size)
	}

	return cf
}

func getBigGoodBase(path string) string {
	base := strings.Replace(strings.Split(filepath.Base(path), ".")[0], "【IRIS OHYAMA】", "", 1)
	return fmt.Sprintf(model.Environment.SE.BigsPathBase, base)
}

func getNormalGoodBase(path string) string {
	base := strings.Replace(strings.Split(filepath.Base(path), ".")[0], "【IRIS OHYAMA】", "", 1)
	return fmt.Sprintf(model.Environment.SE.NormalPathBase, base)
}
