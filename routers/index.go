package routers

import (
	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/iris_auto/middleware"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/process"
)

func Routers(c *augo.Collector) {

	var SplitService *process.SplitFiles = process.NewSplitFiles()
	var ShippService *process.ShippList = process.NewShippList()

	var QCDetail = process.NewDetailQC()
	var WendaQCService *process.WendaQC = process.NewWendaQC(QCDetail)
	var ZhaipeiQCService *process.ZhaipeiQC = process.NewZhaipeiQC(QCDetail)

	c.Use(augo.Recovery(c.Logger))

	Normal_Group := c.Group(absoluteServicePath(augo.GetPathChar()), augo.DeletFiles(), middleware.VerificationPath)
	{
		//spilt-and-export
		Normal_Group.Handler(model.SPILT_AND_EXPORT_MOTHOD, false, SplitService.SplitAndExport(process.OriginSpliteFiles))

		//shipp-list
		Normal_Group.Handler(model.SHIPP_LIST_MOTHOD, false, ShippService.ExportShippList)

		//tripartite-group
		Tripartite_Group := Normal_Group.Group("三方反品")
		{
			//tripartite-spilt-and-export
			Tripartite_Group.HandlerWithVisit(model.TRIPARTITE_SPILT_MOTHOD, SplitService.SplitAndExport(process.TripartiteSplitFiles))
		}
	}

	//QC
	QC_Group := c.Group(absoluteServicePath(augo.GetPathChar()), middleware.VerificationPath, middleware.InitFiles)
	{

		//wenda-qc-table
		QC_Group.HandlerWithVisit(model.WENDA_QC_MOTHOD, WendaQCService.WendaMergeBoxAndExportList)

		//zhaipei-qc-table
		{
			//origin-qc
			QC_Group.HandlerWithVisit(model.ZHAIPEI_QC_MOTHOD, ZhaipeiQCService.OrginZhaipeiQC)

			//third-party-qc
			QC_Group.HandlerWithVisit(model.THIRD_PARTY_QC_MOTHOD, ZhaipeiQCService.ThirdPartyQC)
		}

	}

}
