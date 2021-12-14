package middleware

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/tool"
)

func VerificationPath(c *augo.Context) {

	switch c.Request.Method() {
	case model.SPILT_AND_EXPORT_MOTHOD, model.TRIPARTITE_SPILT_MOTHOD:
		_, _, err := checkFileCount(c.Request, 0, -1)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		s, err := getSourc("", c.Request.Method())
		if err != nil {
			c.AbortWithError(err)
			return
		}

		c.Set(model.SOURCE_KEY, s)

	case model.SHIPP_LIST_MOTHOD:
		xlsxpath, csvpath, err := checkFileCount(c.Request, 1, 1)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		ss, err := getSources(append(xlsxpath, csvpath...), c.Request.Method())
		if err != nil {
			c.AbortWithError(err)
			return
		}

		c.Set(model.SOURCE_KEY, ss[0])
		c.Set(model.CSV_KEY, ss[1])

	case model.WENDA_QC_MOTHOD, model.ZHAIPEI_QC_MOTHOD, model.THIRD_PARTY_QC_MOTHOD:
		xlsxpath, csvpath, err := checkFileCount(c.Request, 1, -1)
		if err != nil {
			c.AbortWithError(err)
			return
		}

		ss, err := getSources(append(xlsxpath, csvpath...), c.Request.Method())
		if err != nil {
			c.AbortWithError(err)
			return
		}

		c.Set(model.SOURCE_KEY, ss[0])
		c.Set(model.CSV_KEY, ss[1])

	default:
		panic(fmt.Sprintf(`"%s" is unknown method`, c.Request.Method()))
	}
}

//確認是否為指定的檔案數量
func checkFileCount(req *augo.Request, csvcount, xlsxcount int) ([]string, []string, error) {

	var (
		csvtotal  int
		xlsxtotal int
		xlsxpaths []string
		csvpath   []string
	)

	for i := 0; i < len(req.Files); i++ {
		file := req.Files[i]
		switch filepath.Ext(file) {
		case ".xlsx":

			//檔案格式為xlsx但檔名為csv時視為csv
			if tool.IsCsvFormat(file) {
				csvpath = append(csvpath, file)
				csvtotal++
				continue
			}

			xlsxpaths = append(xlsxpaths, file)
			xlsxtotal++
		case ".csv":
			csvpath = append(csvpath, file)
			csvtotal++
		case ".DS_Store":
			continue
		default:
			return nil, nil, fmt.Errorf("%s Incorrect file format", file)
		}
	}

	if (csvcount != csvtotal && csvcount != -1) || (csvtotal <= 0 && csvcount == -1) {
		return nil, nil, errors.New("Incorrect number of .csv files entered")
	}
	if (xlsxcount != xlsxtotal && xlsxcount != -1) || (xlsxtotal <= 0 && xlsxcount == -1) {
		return nil, nil, errors.New("Incorrect number of .xlsx files entered")
	}

	return xlsxpaths, csvpath, nil
}

func getSources(paths []string, method string) ([]*excelgo.Sourc, error) {
	var ss []*excelgo.Sourc
	switch method {
	case model.SPILT_AND_EXPORT_MOTHOD:
		if len(paths) <= 0 {
			return nil, errors.New("")
		}

		s, err := getSourc(paths[0], method)
		if err != nil {
			return nil, err
		}

		return append(ss, s), nil

	case model.SHIPP_LIST_MOTHOD:
		if len(paths) != 2 {
			return nil, errors.New("Too many files")
		}

		var csvsourc, commoditysourc excelgo.Sourc
		var s *excelgo.Sourc

		for _, p := range paths {
			if tool.IsCsvFormat(filepath.Base(p)) {
				csvsourc = model.Environment.SL.CsvSourc.Sourc
				s = &csvsourc
			} else {
				commoditysourc = model.Environment.SL.CommoditySourc.Sourc
				s = &commoditysourc
			}

			f, err := excelgo.OpenFile(p)
			if err != nil {
				return nil, err
			}

			if err := s.Init(f); err != nil {
				return nil, err
			}
		}

		if &commoditysourc == nil || &csvsourc == nil {
			return nil, errors.New("Input file type is wrong")
		}
		return append(ss, &commoditysourc, &csvsourc), nil

	case model.WENDA_QC_MOTHOD, model.ZHAIPEI_QC_MOTHOD, model.THIRD_PARTY_QC_MOTHOD:
		var s, csv excelgo.Sourc
		for _, p := range paths {
			if tool.IsCsvFormat(filepath.Base(p)) {
				csv = model.Environment.IDS.CsvSourc.Sourc
				f, err := excelgo.OpenFile(p)
				if err != nil {
					return nil, err
				}

				if err := csv.Init(f); err != nil {
					return nil, err
				}
			}
		}

		if method == model.WENDA_QC_MOTHOD {
			s = model.Environment.IDS.WendaMergeBox.Sourc
		} else {
			s = model.Environment.IDS.ZhaipeiMergeBox.Sourc
		}
		return append(ss, &s, &csv), nil
	}

	return nil, nil
}

func getSourc(path, method string) (*excelgo.Sourc, error) {
	switch method {
	case model.SPILT_AND_EXPORT_MOTHOD:
		s := model.Environment.SE.Sourc
		return &s, nil
	}
	return nil, nil
}
