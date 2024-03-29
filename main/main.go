package main

import (
	"time"

	"github.com/Gaku0607/augo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/formatter"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/routers"
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
