package main

import (
	"strings"
	"time"

	"github.com/Gaku0607/augo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/formatter"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/routers"
	"github.com/xuri/excelize/v2"
)

func main() {

	// if err := iris.WriteDefaultConfig(); err != nil {
	// 	panic(err.Error())
	// }

	go iris.RecoveryPrint()

	if err := iris.InitEnvironment(); err != nil {
		iris.ErrChan <- err
	}

	if err := routers.MakeServiceRouters(); err != nil {
		iris.ErrChan <- err
	}
	//初始化匯出格式
	formatter.InitExportValue()

	//設置環境
	augo.SetSystemVersion(augo.MacOS)
	//設置log標頭
	augo.SetLogTitle("IRIS")

	c := augo.NewCollector()

	//註冊所有服務路由
	routers.Routers(c)

	engine := augo.NewEngine(
		augo.MaxThread(2),
		augo.ScanIntval(time.Second*2),
		augo.SetCollector(c),
		augo.DeleteVisitedIntval(time.Second*time.Duration(model.Delete_Intval)),
	)

	engine.Run()

	engine.Wait()

	if err := engine.Close(); err != nil {
		iris.ErrChan <- err
	}
}

func getCellBgColor(f *excelize.File, sheet, axix string) string {
	styleID, err := f.GetCellStyle(sheet, axix)
	if err != nil {
		return err.Error()
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
