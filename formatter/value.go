package formatter

import (
	"strings"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
)

func InitExportValue() {
	//zhaipei remark
	excelgo.FormatCategory.SetFormatCategory(
		model.ZHAIPEI_QC_MOTHOD,
		model.ZHAIPEI_RETURNS_SHEET,
		"H",
		func(i interface{}) interface{} {
			if remark := i.(string); remark == "" {
				return model.Environment.IDS.ZhaipeiMergeBox.TimeRemark[model.OTHER]
			} else if strings.Contains(remark, model.MORNING) {
				return model.Environment.IDS.ZhaipeiMergeBox.TimeRemark[model.MORNING]
			} else if strings.Contains(remark, model.AFTERNOON) {
				return model.Environment.IDS.ZhaipeiMergeBox.TimeRemark[model.AFTERNOON]
			} else {
				return model.Environment.IDS.ZhaipeiMergeBox.TimeRemark[model.OTHER]
			}
		},
	)

	//zhaipei goods_size
	excelgo.FormatCategory.SetFormatCategory(
		model.ZHAIPEI_QC_MOTHOD,
		model.ZHAIPEI_RETURNS_SHEET,
		"J",
		func(i interface{}) interface{} {
			if val, exit := model.Environment.IDS.ZhaipeiMergeBox.Sizelist[i.(string)]; exit {
				return val
			} else {
				return "null"
			}
		},
	)
}
