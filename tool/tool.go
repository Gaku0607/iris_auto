package tool

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Gaku0607/iris_auto/model"
	"github.com/xuri/excelize/v2"
)

func ErrMsgs(assertion bool, msg string) {
	if !assertion {
		panic(msg)
	}
}

//查看字符串是否非數字組成
func IsNumeric(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

func IsCsvFormat(filename string) bool {
	b, err := regexp.MatchString(`(csv|CSV){1}`, filename)
	if err != nil {
		return false
	}
	return b
}

func IsExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func GetUniqueCode(origincode string) string {
	origincode = strings.TrimSpace(origincode)
	//取注文番號前14碼
	if (strings.Contains(origincode, "-") || origincode[len(origincode)-1] == 'A') && len(origincode) >= 14 {
		return origincode[:14]
	} else {
		return origincode
	}
}

func IsAnnotaion(str string) bool {
	if str == "" {
		return false
	}
	return uint8(str[0]) == []uint8(`#`)[0]
}

func IsTripartiteQCMethod(filename string) bool {
	return strings.Contains(filepath.Base(filename), model.Environment.TF.TripartiteInputFormat)
}

//獲取ＣELL顏色
func GetCellBgColor(f *excelize.File, sheet, axix string) string {
	styleID, err := f.GetCellStyle(sheet, axix)
	if err != nil {
		return ""
	}
	fillID := *f.Styles.CellXfs.Xf[styleID].FillID
	fgColor := f.Styles.Fills.Fill[fillID].PatternFill.FgColor

	if fgColor == nil {
		return ""
	}

	if fgColor.Theme != nil {
		children := f.Theme.ThemeElements.ClrScheme.Children
		if *fgColor.Theme < 4 {
			dklt := map[int]string{
				0: children[1].SysClr.LastClr,
				1: children[0].SysClr.LastClr,
				2: *children[3].SrgbClr.Val,
				3: *children[2].SrgbClr.Val,
			}
			return strings.TrimPrefix(
				excelize.ThemeColor(dklt[*fgColor.Theme], fgColor.Tint), "FF")
		}
		srgbClr := *children[*fgColor.Theme].SrgbClr.Val
		return strings.TrimPrefix(excelize.ThemeColor(srgbClr, fgColor.Tint), "FF")
	}
	return strings.TrimPrefix(fgColor.RGB, "FF")
}
