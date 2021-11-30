package iris

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Gaku0607/iris_auto/model"
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
	return json.Unmarshal(data, &model.Environment.SE)
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
