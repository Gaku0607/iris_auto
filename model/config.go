package model

import (
	"github.com/Gaku0607/excelgo"
)

//**************************************
//***************EnvPaths***************
//**************************************

var Delete_Intval int

var EnvironmentDir string //配置文件資料夾地址

var Services_Dir string //服務資料夾

var Result_Path string //完成時返回地址

var EnvPath = "/Users/gaku/Documents/GitHub/iris_auto/main/.env"

//**************************************
//***************EnvSources*************
//**************************************

var Environment SystemParms

type SystemParms struct {
	SE  SpiltAndExportParms  `json:"spilt_and_export_parms"` //分割匯出 所需要的參數
	SL  ShippingListParms    `json:"shipping_list_parms"`    //出倉單所需要的參數
	IDS ImportDocumentsParms `json:"import_documents_parms"` //匯入當月分文件時所需要的參數
}

type ShippingListParms struct {
	CsvSourc          `json:"csv_sourc"`
	CommoditySourc    `json:"commodity_sourc"`
	ShippListPosition `json:"shipp_list_position"`
	FileFormatPath    string `json:"file_format_path"`
	OutputFileFormat  string `json:"output_file_format"`
}

type CsvSourc struct {
	excelgo.Sourc
	CodeSpan       string `json:"code_span"`
	AreaSpan       string `json:"area_span"`
	TelSpan        string `json:"tel_span"`
	AddrSpan       string `json:"addr_span"`
	RecipientSpna  string `json:"recipient_span"`
	ItemsTotalSpan string `json:"items_total_span"`
}

type CommoditySourc struct {
	excelgo.Sourc
	CodeSpan    string `json:"code_span"`
	JANCodeSpan string `json:"jan_code_span"`
	DateIndex   int    `json:"date_index"`
}

type ShippListPosition struct {
	DataPosition      int    `json:"data_position"`       //輸入資料位置
	RecipientPosition string `json:"recipient_postition"` //接收人位置
	AddrPosition      string `json:"addr_position"`       //地址位置
	TelPosition       string `json:"tel_position"`        // 電話位置
	OrderNumPosition  string `json:"order_num_position"`  // 單號位置
	PartyPosition     string `json:"party_position"`      //負責人位置
	DatePosition      string `json:"date_position"`       //日期位置
	TotalPosition     string `json:"total_position"`      //總數位置

	TotalIndex         int    `json:"total_index"`
	PackingMethodIndex int    `json:"packing_method_index"`
	PackingMethod      string `json:"packing_method"` //裝箱方式
}

type SpiltAndExportParms struct {
	excelgo.Sourc
	SlicerParms
	BigsPathBase   string `json:"big_goods_path_base"`
	NormalPathBase string `json:"normal_goods_path_base"`
}

type SlicerParms struct {
	Identify     interface{} `json:"identify"`
	SizeSpan     string      `json:"size_span"`
	Size         int         `json:"size"`
	ContainsSpan string      `json:"contains_span"`
	Contains     []string    `json:"contains"`
}

type ImportDocumentsParms struct {
	CsvSourc           `json:"csv_sourc"`
	ToBeConfirmedSheet string `json:"to_be_confirmed_sheet"` //待確認頁籤
	BoxTotalFormula    string `json:"box_total_formula"`
	BoxCountSpan       string `json:"box_count_span"`
	RemarkSpan         string `json:"remark_span"`      //備註
	PostalCodeSpan     string `json:"postal_code_span"` // 郵遞區號
	TotalSpan          string `json:"total_span"`       //商品數
	ItemCodeSpan       string `json:"item_code_span"`   //商品6碼
	QuantitySpan       string `json:"quantity_span"`    //商品入數
	UniqueCodeSpan     string `json:"unique_code_span"` //客戶訂單碼 （注文番號）

	//宅配通所使用併箱格式
	ZhaipeiMergeBox struct {
		ThirdPartySort           map[string]excelgo.Order `json:"third_party_sort"` //第三方排序
		ThirdPartyMasterFileBase string                   `json:"third_party_master_file_base"`

		excelgo.Sourc     `json:"sourc"`
		NewHeaders        []string          `json:"new_headers"` //宅單所需要的Headers
		NewCols           []*excelgo.Col    `json:"new_cols"`    //宅配主檔新欄位
		ArrivalRemarkSpan string            `json:"arrival_remark_span"`
		TimeRemark        map[string]int    `json:"time_remark"` //送貨備註
		SizeSpan          string            `json:"size_span"`
		Sizelist          map[string]string `json:"size_list"` //尺寸表
		JanCodeSpan       string            `json:"jan_code_span"`
		MasterFileBase    string            `json:"master_file_base"` //總檔的檔名格式
	} `json:"zhaipei_merge_box,omitempty"`

	//穩達所使用的併箱格式
	WendaMergeBox struct {
		excelgo.Sourc            `json:"sourc"`
		NewCols                  []*excelgo.Col `json:"new_cols"`    //穩達主檔新欄位
		NewHeaders               []string       `json:"new_headers"` //穩達主檔新欄位的標頭
		CountStrSpan             string         `json:"count_str_span"`
		DeliveryPath             string         `json:"delivery_path"`     //穩達所使用的宅配文件地址
		WendaFormatPath          string         `json:"wenda_format_path"` //回傳穩達格式的文件地址
		CustomerCode             string         `json:"coustomer_code"`    //客戶代碼 （固定值）
		CustomerCodeTCol         string         `json:"coustomer_code_tcol"`
		DeliveryInstallation     int            `json:"delivery_installation"` //宅配安裝 （固定值）
		DeliveryInstallationTCol string         `json:"delivery_installation_tcol"`
		IsSelfCollection         string         `json:"is_self_collection"` //是否自取 (固定值)
		IsSelfCollectionTCol     string         `json:"is_self_collection_tcol"`
		MasterFileBase           string         `json:"master_file_base"` //總檔的檔名格式
	} `json:"wanda_merge_box,omitempty"`
}
