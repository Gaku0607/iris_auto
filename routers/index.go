package routers

import (
	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/iris_auto/middleware"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/process"
)

func Routers(c *augo.Collector) {

	c.Use(middleware.VerificationPath)

	{
		//spilt-and-export
		c.Handler(absoluteServicePath(model.SPILT_AND_EXPORT_MOTHOD), process.NewSplitFiles().SplitAndExport)

		//shipp-list
		c.Handler(absoluteServicePath(model.SHIPP_LIST_MOTHOD), process.NewShippList().ExportShippList)

		//QC
		QC_Group := c.Group(absoluteServicePath(augo.GetPathChar()), middleware.InitSourcFiles)
		{
			qc_detail := process.NewDetailQC()

			//wenda-qc-table
			QC_Group.Handler(model.WENDA_QC_MOTHOD, process.NewWendaQC(qc_detail).WendaMergeBoxAndExportList)

			//zhaipei-qc-table
			zhaipei := process.NewZhaipeiQC(qc_detail)
			{
				//origin-qc
				QC_Group.Handler(model.ZHAIPEI_QC_MOTHOD, zhaipei.OrginZhaipeiQC)

				//third-party-qc
				QC_Group.Handler(model.THIRD_PARTY_QC_MOTHOD, zhaipei.ThirdPartyQC)
			}

		}
	}

}
