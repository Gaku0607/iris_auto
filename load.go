package iris

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Gaku0607/augo"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/tool"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
)

// \\\\10.212.38.250\\Focus 資材\\ILC\\17.愛麗思歐雅瑪(IRIS OHYAMA)\\iris_system\\model\\出倉單.xlsx
//  \\10.212.38.250\Focus 資材\ILC\17.愛麗思歐雅瑪(IRIS OHYAMA)\iris_system\main\.env`

//從配置文件中加載所有環境變數
func InitEnvironment() error {
	var err error
	if err := godotenv.Load(model.EnvPath); err != nil {
		return err
	}
	//配置文件地址
	if err := loadPath(); err != nil {
		return err
	}

	//刪除間格時間
	Intval := os.Getenv("delete_intval")
	if Intval == "" {
		return errors.New("DeleteIntval is not exist")
	}

	model.Delete_Intval, err = strconv.Atoi(Intval)
	if err != nil {
		return err
	}

	if err := loadEnv(); err != nil {
		return err
	}

	//清除出倉單舊歷程
	if err := clearShippListHistory(); err != nil {
		return err
	}

	return loadDeliverAria()
}

func loadPath() error {
	model.EnvironmentDir = os.Getenv("environment_dir")
	if model.EnvironmentDir == "" {
		return errors.New("EnvironmentDir is not exist")
	}

	model.Services_Dir = os.Getenv("services_dir")
	if model.Services_Dir == "" {
		return errors.New("ServicesDir is not exist")
	}

	model.Result_Path = os.Getenv("result_path")
	if model.Result_Path == "" {
		return errors.New("ResultPath is not exist")
	}

	return nil
}

func loadEnv() error {

	data, err := load(model.QC_CSV_BASE)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &model.Environment.IDS); err != nil {
		return err
	}

	//穩達
	if data, err = load(model.WENDA_QC_BASE); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &model.Environment.IDS.WendaMergeBox); err != nil {
		return err
	}

	//宅配通
	if data, err = load(model.ZHAIPEI_QC_BASE); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &model.Environment.IDS.ZhaipeiMergeBox); err != nil {
		return err
	}

	//出倉單
	if data, err = load(model.SHIPP_LIST_BASE); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &model.Environment.SL); err != nil {
		return err
	}

	//切割文件
	if data, err = load(model.SPILT_AND_EXPORT_BASE); err != nil {
		return err
	}
	if err = json.Unmarshal(data, &model.Environment.SE); err != nil {
		return err
	}

	//三方表單
	if data, err = load(model.TRIPARTITE_FORM_BASE); err != nil {
		return err
	}
	if err = json.Unmarshal(data, &model.Environment.TF); err != nil {
		return err
	}

	if err = initTripartiteStatusForm(); err != nil {
		return err
	}

	return nil
}

func load(path string) ([]byte, error) {

	file, err := os.OpenFile(filepath.Join(model.EnvironmentDir, path), os.O_RDWR|os.O_CREATE, os.ModeAppend|os.ModePerm)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	return ioutil.ReadAll(file)
}

func loadDeliverAria() error {

	const (
		CityCol int = iota
		PartitionCol
		PostalCodeCol
		AreaCodeCol
		PlaceCol
	)

	file, err := excelize.OpenFile(model.Environment.IDS.WendaMergeBox.DeliveryPath)
	if err != nil {
		return err
	}

	rows, err := file.GetRows(file.GetSheetName(0))
	if err != nil {
		return err
	}

	for _, row := range rows {
		delivery := &model.Delivery{
			City:       row[CityCol],
			Partition:  row[PartitionCol],
			PostalCode: row[PostalCodeCol],
			AreaCode:   row[AreaCodeCol],
			Place:      row[PlaceCol],
		}
		model.DeliveryArea[row[PostalCodeCol]] = delivery
	}

	return nil
}

//初始化三方表單狀態
func initTripartiteStatusForm() error {
	model.TripartiteStatus = model.Environment.TF.TripartiteStatusList
	if model.TripartiteStatus == nil {
		return errors.New("Failed to initialize TripartiteStatusForm")
	}
	model.TripartiteStatus = append(model.TripartiteStatus, model.TRIPARTITE_STATUS_NULL)
	return nil
}

//當出倉單的歷程時間比當日還舊時進行清除
func clearShippListHistory() error {

	annotation := []byte(`##格式為 ex:19960607=14 , "="左側為日期 "="右側為單號`)

	date_env, err := os.OpenFile(model.Environment.SL.HistoryEnvPath, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	olddata, err := ioutil.ReadAll(date_env)
	if err != nil {
		return err
	}

	//每條歷程
	rows := bytes.Split(olddata, []byte(augo.GetNewLine()))

	//當日時間
	todaystr := time.Now().Format("20060102")
	today, _ := strconv.Atoi(todaystr)

	var (
		newrows      [][]byte //用於儲存新的歷程
		historycount int      //儲存所有歷程數量
	)

	for _, row := range rows {

		//空行
		if len(row) == 0 {
			continue
		}
		//註解
		if tool.IsAnnotaion(string(row)) {
			continue
		}

		historycount++

		linebyte := bytes.SplitN(row, []byte("="), 2)
		if len(linebyte) != 2 {
			continue
		}

		datestr := string(linebyte[0])
		date, _ := strconv.Atoi(datestr)
		if today > date {
			continue
		}

		newrows = append(newrows, row)
	}

	if len(newrows) == historycount {
		return nil
	}

	if err := date_env.Truncate(0); err != nil {
		return err
	}

	date_env.Seek(0, 0)

	newdata := bytes.Join(newrows, []byte(augo.GetNewLine()))
	temp := newdata
	linelen := len([]byte(augo.GetNewLine()))
	newdata = make([]byte, len(newdata)+len(annotation)+(linelen*2))

	copy(newdata, annotation)
	copy(newdata[len(annotation):], augo.GetNewLine())
	copy(newdata[len(annotation)+linelen:], temp)
	copy(newdata[len(annotation)+len(temp)+linelen:], augo.GetNewLine())

	if _, err := date_env.Write(newdata); err != nil {
		return err
	}

	return nil
}
