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

	Normal_Group := c.Group(model.Services_Dir, augo.DeletFiles(), middleware.VerificationPath)
	{
		//spilt-and-export
		Normal_Group.Handler(model.SPILT_AND_EXPORT_MOTHOD, false, SplitService.SplitAndExport(process.OriginSpliteFiles))

		//shipp-list
		Normal_Group.Handler(model.SHIPP_LIST_MOTHOD, false, ShippService.ExportShippList)

	}

	Visit_Group := c.Group(model.Services_Dir, middleware.VerificationPath)
	{
		//tripartite-group
		Tripartite_Group := Visit_Group.Group("三方反品")
		{
			//tripartite-spilt-and-export
			Tripartite_Group.HandlerWithVisit(model.TRIPARTITE_SPILT_MOTHOD, SplitService.SplitAndExport(process.TripartiteSplitFiles))

			//tripartite-qc-group
			Tripartite_QC_Group := Tripartite_Group.Group(augo.GetPathChar(), middleware.InitTripartiteForm(), middleware.InitQCSourceFiles)
			{
				//tripartite-zhaipei-qc
				Tripartite_QC_Group.HandlerWithVisit(model.TRIPARTITE_ZHAIPEI_QC_MOTHOD, ZhaipeiQCService.TripartiteQC)
			}

		}

		QC_Group := Visit_Group.Group(augo.GetPathChar(), middleware.InitQCSourceFiles)
		{

			//wenda-qc-table
			QC_Group.HandlerWithVisit(model.WENDA_QC_MOTHOD, WendaQCService.WendaMergeBoxAndExportList)

			//zhaipei-origin-qc
			QC_Group.HandlerWithVisit(model.ZHAIPEI_QC_MOTHOD, ZhaipeiQCService.ZhaipeiQC(true))

			//zhaipei-third-party-qc
			QC_Group.HandlerWithVisit(model.THIRD_PARTY_QC_MOTHOD, ZhaipeiQCService.ZhaipeiQC(false))

		}

	}

}
