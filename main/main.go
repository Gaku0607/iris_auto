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

	//設置環境
	augo.SetSystemVersion(augo.MacOS)
	//設置log標頭
	augo.SetLogTitle("IRIS")

	//初始化匯出格式
	formatter.InitExportValue()
	formatter.InitExportFormula()

	c := augo.DefautCollector(
		augo.ResultLogKey(
			func(c *augo.Context) augo.LogKey {
				key := make(augo.LogKey)
				for k, v := range c.Keys {
					if !(k == model.SOURCE_KEY || k == model.CSV_KEY) {
						key[k] = v
					}
				}
				return key
			}),
	)

	//註冊所有服務路由
	routers.Routers(c)

	engine := augo.NewEngine(
		augo.MaxThread(2),
		augo.ScanIntval(time.Second*2),
		augo.SetCollector(c),
	)

	engine.Run()

	engine.Wait()

	if err := engine.Close(); err != nil {
		panic(err.Error())
	}
}
