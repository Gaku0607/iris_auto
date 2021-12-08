package routers

import (
	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/iris_auto/middleware"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/process"
)

func Routers(c *augo.Collector) {

	c.Use(augo.Recovery(c.Logger))

	Normal_Group := c.Group(absoluteServicePath(augo.GetPathChar()), augo.DeletFiles(), middleware.VerificationPath)
	{
		//spilt-and-export
		Normal_Group.Handler(model.SPILT_AND_EXPORT_MOTHOD, false, process.NewSplitFiles().SplitAndExport)

		//shipp-list
		Normal_Group.Handler(model.SHIPP_LIST_MOTHOD, false, process.NewShippList().ExportShippList)

		//tripartite-form
		Normal_Group.Handler(model.TRIPARTITE_FORM_MOTHOD, false, process.NewTripartiteForm().TripartiteForm)
	}

	//QC
	QC_Group := c.Group(absoluteServicePath(augo.GetPathChar()), middleware.VerificationPath, middleware.InitFiles)
	{
		qc_detail := process.NewDetailQC()

		//wenda-qc-table
		QC_Group.HandlerWithVisit(model.WENDA_QC_MOTHOD, process.NewWendaQC(qc_detail).WendaMergeBoxAndExportList)

		//zhaipei-qc-table
		zhaipei := process.NewZhaipeiQC(qc_detail)
		{
			//origin-qc
			QC_Group.HandlerWithVisit(model.ZHAIPEI_QC_MOTHOD, zhaipei.OrginZhaipeiQC)

			//third-party-qc
			QC_Group.HandlerWithVisit(model.THIRD_PARTY_QC_MOTHOD, zhaipei.ThirdPartyQC)
		}

	}

}
