package iris

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/Gaku0607/excelgo"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/joho/godotenv"
)

//寫入默認的配置文件
func WriteDefaultConfig() error {

	if err := godotenv.Load(model.EnvPath); err != nil {
		return err
	}

	if err := loadPath(); err != nil {
		return err
	}

	s := &model.SystemParms{}

	//分割以及匯出 （分專車宅配）
	if err := spilt_and_export_parms(s.SE); err != nil {
		return err
	}

	//出倉單
	if err := shipping_list_parms(s.SL); err != nil {
		return err
	}

	//三方表單
	if err := tripartite_form_parms(&s.TF); err != nil {
		return err
	}

	//QC表單
	if err := qc_form_parms(&s.IDS); err != nil {
		return err
	}

	return nil
}

func writeConfig(data interface{}, base string) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(model.EnvironmentDir, base), b, 0777)
}

func spilt_and_export_parms(se model.SpiltAndExportParms) error {
	sumcol := excelgo.NewCol("合計数")
	sumcol.Numice = excelgo.Numice{IsNumice: true, IsSum: true}

	momo := excelgo.NewSourc(
		model.ORIGIN_SHEET,
		excelgo.NewCol("箱サイズ合計"),
		excelgo.NewCol("注文番号"),
		sumcol,
		excelgo.NewCol("運送便名称"),
	)
	momo.SpanSorts = []excelgo.SpanSort{{Span: "注文番号", Order: excelgo.PositiveOrder}, {Span: "納品先名称", Order: excelgo.PositiveOrder}}

	sp := model.SlicerParms{}
	sp.Identify = "momo第三方指示"
	sp.SizeSpan = "箱サイズ合計"
	sp.Size = 160
	sp.ContainsSpan = "運送便名称"
	sp.Contains = []string{"台湾日通(MOMO指定)"}

	se = model.SpiltAndExportParms{
		Sourc:          *momo,
		SlicerParms:    sp,
		BigsPathBase:   "%s-專車.xlsx",
		NormalPathBase: "%s-宅配.xlsx",
	}
	return writeConfig(&se, model.SPILT_AND_EXPORT_BASE)
}

func shipping_list_parms(sl model.ShippingListParms) error {

	boxtotal := excelgo.NewCol("入数")
	boxtotal.Numice = excelgo.Numice{IsNumice: true}
	boxtotal.TCol = []*excelgo.TargetCol{excelgo.NewTCol(model.SHIPP_LIST_MOTHOD, "", "H")}

	shipptotal := excelgo.NewCol("出荷予定総バラ数")
	shipptotal.Numice = excelgo.Numice{IsNumice: true}
	shipptotal.TCol = []*excelgo.TargetCol{excelgo.NewTCol(model.SHIPP_LIST_MOTHOD, "", "J")}

	warecodecol := excelgo.NewCol("商品コード")
	warecodecol.TCol = []*excelgo.TargetCol{excelgo.NewTCol(model.SHIPP_LIST_MOTHOD, "", "B")}

	warecol := excelgo.NewCol("商品名称1ST")
	warecol.TCol = []*excelgo.TargetCol{excelgo.NewTCol(model.SHIPP_LIST_MOTHOD, "", "C")}

	csv := excelgo.NewSourc(
		model.ORIGIN_SHEET,
		warecodecol,
		warecol,
		boxtotal,
		shipptotal,
		excelgo.NewCol("エリア"),
		excelgo.NewCol("配送先住所1"),
		excelgo.NewCol("配送先名称1ST"),
		excelgo.NewCol("配送先電話番号"),
	)

	commodity := excelgo.NewSourc(
		model.ORIGIN_SHEET,
		excelgo.NewCol("商品コード"),
		excelgo.NewCol("JAN コード"),
	)

	sl = model.ShippingListParms{
		CsvSourc: model.CsvSourc{
			Sourc:         *csv,
			AreaSpan:      "エリア",
			AddrSpan:      "配送先住所1",
			RecipientSpna: "配送先名称1ST",
			TelSpan:       "配送先電話番号",
			CodeSpan:      "商品コード",
		},
		CommoditySourc: model.CommoditySourc{
			Sourc:       *commodity,
			CodeSpan:    "商品コード",
			JANCodeSpan: "JAN コード",
			DateIndex:   1,
		},
		ShippListPosition: model.ShippListPosition{
			DataPosition:      10,
			RecipientPosition: "A6",
			AddrPosition:      "A7",
			TelPosition:       "A8",
			OrderNumPosition:  "K6",
			PartyPosition:     "K8",
			DatePosition:      "K7",
			TotalPosition:     "J35",

			TotalIndex:         3,
			PackingMethodIndex: 2,
			PackingMethod:      "箱",
		},
		FileFormatPath:   `/Users/gaku/IRIS系統測試檔案/出倉單/個例/出倉單.xlsx`,
		OutputFileFormat: `%s-iris出倉單_%s.xlsx`,
		HistoryEnvPath:   "/Users/gaku/Documents/GitHub/iris_auto/config/shipplist_history.env",
	}
	return writeConfig(&sl, model.SHIPP_LIST_BASE)
}

//***************************************
//****************QC FORM****************
//***************************************

func qc_form_parms(ids *model.ImportDocumentsParms) error {

	//QC表單格式
	if err := import_docments_parms(ids); err != nil {
		return err
	}

	//宅配QC格式
	if err := zhaipei_qc_parms(ids); err != nil {
		return err
	}

	//穩達QC格式
	if err := wenda_qc_parms(ids); err != nil {
		return err
	}
	return nil
}

func import_docments_parms(ids *model.ImportDocumentsParms) error {
	//待確認頁籤
	ids.ToBeConfirmedSheet = "Sheet2"
	//csv的Sourc
	csv := ids.CsvSourc
	{
		csv.CodeSpan = "商品コード"
		csv.AreaSpan = "エリア"
		csv.ItemsTotalSpan = "出荷予定総バラ数"
		csv.Sourc = *excelgo.NewSourc(
			model.ORIGIN_SHEET,
			excelgo.NewCol(csv.CodeSpan),
			excelgo.NewCol(csv.AreaSpan),
			excelgo.NewCol(csv.ItemsTotalSpan),
		)
	}
	ids.CsvSourc = csv
	ids.BoxTotalFormula = `=SUM(E%d:E%d)`
	ids.BoxCountSpan = "箱數"
	ids.RemarkSpan = "梱包指示"
	ids.PostalCodeSpan = "届先郵便番号"
	ids.TotalSpan = "合計数"
	ids.ItemCodeSpan = "商品コード"
	ids.QuantitySpan = "入数"
	ids.UniqueCodeSpan = "注文番号"

	return writeConfig(&ids, model.QC_CSV_BASE)
}

func zhaipei_qc_parms(ids *model.ImportDocumentsParms) error {
	//宅配通PARMS
	ids.ZhaipeiMergeBox.NewHeaders = []string{"序號", "訂單編號-1", "箱數", "倉別", "尾碼"}
	//輸出文件格式
	ids.ZhaipeiMergeBox.MasterFileBase = "%s-宅配通QC.xlsx"
	//第三方排序
	ids.ZhaipeiMergeBox.ThirdPartySort = []excelgo.SpanSort{{Span: "受注番号", Order: excelgo.PositiveOrder}}
	//第三方檔名格式
	ids.ZhaipeiMergeBox.ThirdPartyMasterFileBase = "%s-第三方宅配QC.xlsx"
	//尺寸對照表
	ids.ZhaipeiMergeBox.SizeSpan = "箱サイズ合計"
	ids.ZhaipeiMergeBox.Sizelist = map[string]string{
		"60":  "01",
		"90":  "02",
		"120": "03",
		"150": "04",
		"180": "05",
	}
	//到貨對照表
	ids.ZhaipeiMergeBox.ArrivalRemarkSpan = "出荷指示備考"
	ids.ZhaipeiMergeBox.TimeRemark = map[string]int{
		"早上":    1,
		"下午":    2,
		"other": 0,
	}

	//NewCol
	{
		ids.ZhaipeiMergeBox.NewCols = []*excelgo.Col{
			excelgo.NewCol("訂單編號-1"),
			excelgo.NewCol("序號"),
			excelgo.NewCol("倉別"),
			excelgo.NewCol("尾碼"),
			excelgo.NewCol("箱數"),
		}
	}

	//宅配通的Sourc
	{

		tcolfn := func(sheet, header string) *excelgo.TargetCol {
			return excelgo.NewTCol(model.ZHAIPEI_QC_MOTHOD, sheet, header)
		}

		totalcol := excelgo.NewCol(ids.TotalSpan)
		totalcol.Numice = excelgo.Numice{IsNumice: true}

		quantitycol := excelgo.NewCol(ids.QuantitySpan)
		quantitycol.Numice = excelgo.Numice{IsNumice: true}

		itemcodecol := excelgo.NewCol(ids.ItemCodeSpan)

		remarkcol := excelgo.NewCol(ids.RemarkSpan)

		ids.ZhaipeiMergeBox.JanCodeSpan = "JAN コード"
		jancodecol := excelgo.NewCol(ids.ZhaipeiMergeBox.JanCodeSpan)

		namecol := excelgo.NewCol("納品先名称")
		namecol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "B")}

		telcol := excelgo.NewCol("届先電話番号")
		telcol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "C")}

		postalcol := excelgo.NewCol("届先郵便番号")
		postalcol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "D")}

		addrcol := excelgo.NewCol("納品先住所")
		addrcol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "E")}

		deliverycol := excelgo.NewCol("出荷指示備考")
		deliverycol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "F"), tcolfn(model.ZHAIPEI_RETURNS_SHEET, "H")}

		datecol := excelgo.NewCol("納品日")
		datecol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "G")}

		uniquecodecol := excelgo.NewCol(ids.UniqueCodeSpan)
		uniquecodecol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "I")}

		boxsizecol := excelgo.NewCol("箱サイズ合計")
		boxsizecol.TCol = []*excelgo.TargetCol{tcolfn(model.ZHAIPEI_RETURNS_SHEET, "J")}

		zhaipeisourc := excelgo.NewSourc(
			model.ORIGIN_SHEET,
			totalcol,
			quantitycol,
			itemcodecol,
			remarkcol,
			uniquecodecol,
			namecol,
			telcol,
			postalcol,
			datecol,
			addrcol,
			deliverycol,
			boxsizecol,
			jancodecol,
		)
		zhaipeisourc.SpanSorts = []excelgo.SpanSort{{Span: "訂單編號-1", Order: excelgo.PositiveOrder}}
		zhaipeisourc.Formulas = excelgo.Formulas{excelgo.NewFormula(model.ZHAIPEI_QC_MOTHOD, fmt.Sprintf("=%s!C", model.ZHAIPEI_QC_SHEET)+"%d", model.ZHAIPEI_RETURNS_SHEET, "A")}
		ids.ZhaipeiMergeBox.Sourc = *zhaipeisourc
	}

	return writeConfig(&ids.ZhaipeiMergeBox, model.ZHAIPEI_QC_BASE)
}

func wenda_qc_parms(ids *model.ImportDocumentsParms) error {

	tcolfn := func(sheet, header string) *excelgo.TargetCol {
		return excelgo.NewTCol(model.WENDA_QC_MOTHOD, sheet, header)
	}

	ordercol := excelgo.NewCol("宅配單號")
	ordercol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_RETURNS_SHEET, "B"), tcolfn(model.WENDA_DRIVER_SHEET, "C")}
	boxcuntcol := excelgo.NewCol("箱數")

	countstr := excelgo.NewCol("件數")

	numcol := excelgo.NewCol("序號")
	numcol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_DRIVER_SHEET, "A")}
	ids.WendaMergeBox.NewCols = []*excelgo.Col{
		countstr,
		boxcuntcol,
		ordercol,
		numcol,
	}
	//穩達宅配地區配置文件位置
	ids.WendaMergeBox.DeliveryPath = "/Users/gaku/IRIS系統測試檔案/穩達匯出/wendadelivery.xlsx"
	//穩達匯出文件格式地址
	ids.WendaMergeBox.WendaFormatPath = "/Users/gaku/IRIS系統測試檔案/穩達匯出/wendaformat.xlsx"
	//穩達客代
	ids.WendaMergeBox.CustomerCodeTCol = "A"
	ids.WendaMergeBox.CustomerCode = "NT"
	//輸出文件格式
	ids.WendaMergeBox.MasterFileBase = "%s-穩達QC.xlsx"
	//NewHeaders
	ids.WendaMergeBox.NewHeaders = []string{"序號", "宅配單號", "件數", "箱數", "區域碼", "地點", "區域"}
	ids.WendaMergeBox.CountStrSpan = "件數"
	//自取
	ids.WendaMergeBox.IsSelfCollection = "N"
	ids.WendaMergeBox.IsSelfCollectionTCol = "K"
	//宅配安裝
	ids.WendaMergeBox.DeliveryInstallation = 0
	ids.WendaMergeBox.DeliveryInstallationTCol = "L"

	//穩達的Sourc
	{
		totalcol := excelgo.NewCol(ids.TotalSpan)
		totalcol.Numice = excelgo.Numice{IsNumice: true}

		quantitycol := excelgo.NewCol(ids.QuantitySpan)
		quantitycol.Numice = excelgo.Numice{IsNumice: true}

		remarkcol := excelgo.NewCol(ids.RemarkSpan)

		itemcodecol := excelgo.NewCol(ids.ItemCodeSpan)

		uniquecodecol := excelgo.NewCol(ids.UniqueCodeSpan)
		uniquecodecol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_DRIVER_SHEET, "B")}

		postalcol := excelgo.NewCol(ids.PostalCodeSpan)
		postalcol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_RETURNS_SHEET, "I")}

		//匯出指定欄位回傳穩達檔案
		namecol := excelgo.NewCol("納品先名称")
		namecol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_RETURNS_SHEET, "F"), tcolfn(model.WENDA_DRIVER_SHEET, "D")}

		telcol := excelgo.NewCol("届先電話番号")
		telcol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_RETURNS_SHEET, "G")}
		addrcol := excelgo.NewCol("納品先住所")
		addrcol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_RETURNS_SHEET, "J")}
		deliverycol := excelgo.NewCol("出荷指示備考")
		deliverycol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_RETURNS_SHEET, "M")}
		jancodecol := excelgo.NewCol("JAN コード")
		jancodecol.TCol = []*excelgo.TargetCol{tcolfn(model.WENDA_RETURNS_SHEET, "N")}

		wendasourc := excelgo.NewSourc(
			model.ORIGIN_SHEET,
			totalcol,
			quantitycol,
			remarkcol,
			uniquecodecol,
			postalcol,
			itemcodecol,
			namecol,
			telcol,
			addrcol,
			deliverycol,
			jancodecol,
		)

		wendasourc.SpanSorts = []excelgo.SpanSort{
			{Span: "JAN コード", Order: excelgo.ReverseOrder},
			{Span: "區域", Order: excelgo.ReverseOrder},
			{Span: "注文番号", Order: excelgo.ReverseOrder},
		}
		wendasourc.Formulas = excelgo.Formulas{
			excelgo.NewFormula(model.WENDA_QC_MOTHOD, fmt.Sprintf("=%s!D", model.WENDA_QC_SHEET)+"%d", model.WENDA_DRIVER_SHEET, "E"),
			excelgo.NewFormula(model.WENDA_QC_MOTHOD, fmt.Sprintf("=%s!D", model.WENDA_QC_SHEET)+"%d", model.WENDA_RETURNS_SHEET, "O"),
		}
		ids.WendaMergeBox.Sourc = *wendasourc
	}
	return writeConfig(&ids.WendaMergeBox, model.WENDA_QC_BASE)
}

//**************************************
//**************Tripartite**************
//**************************************

func tripartite_form_parms(tf *model.TripartiteFormParms) error {

	//三方表單格式
	if err := tripartite_file_parms(tf); err != nil {
		return err
	}

	//三方宅配ＱＣ
	if err := tripartite_zhaipei_qc_parms(tf); err != nil {
		return err
	}

	//三方回傳
	if err := triparite_return_parms(tf); err != nil {
		return err
	}
	return nil
}

func tripartite_file_parms(tf *model.TripartiteFormParms) error {

	tff := model.TripartiteFile{}
	tff.SheetName = "處理中"
	tff.StatusTagCol = "C"
	tff.TripartiteInputFormat = "三方連動"
	tff.UniqueCodeSpan = "訂單編號"
	tf.TripartiteFile = tff

	return writeConfig(&tff, model.TRIPARTITE_FORM_BASE)
}

func triparite_return_parms(tf *model.TripartiteFormParms) error {

	trf := model.TripartiteReturn{}
	trf.GoodsReturnDateSpan = "商品退倉日期"
	trf.GoodsCodeSpan = "JANCODE"
	trf.TotalSpan = "實收數量"

	sourc := excelgo.NewSourc(
		tf.SheetName,
		excelgo.NewCol(trf.GoodsReturnDateSpan),
		excelgo.NewCol(trf.GoodsCodeSpan),
		excelgo.NewCol(trf.TotalSpan),
		excelgo.NewCol(tf.UniqueCodeSpan),
	)

	trf.Sourc = *sourc

	return writeConfig(&trf, model.TRIPARTITE_RETURN_BASE)
}

func tripartite_zhaipei_qc_parms(tf *model.TripartiteFormParms) error {
	tqc := model.TripartiteQC{}
	{
		tqc.TripartiteStatusList = []string{"momo第三方交換", "補寄商品", "來回件"}
		tqc.DateSpan = "聯絡日"
		tqc.GoodsStatusSpan = "作業指示"

		sourc := excelgo.NewSourc(
			tf.SheetName,
			excelgo.NewCol(tqc.DateSpan),
			excelgo.NewCol(tqc.GoodsStatusSpan),
			excelgo.NewCol(tf.UniqueCodeSpan),
		)
		sourc.SpanSorts = []excelgo.SpanSort{{Span: "訂單編號-1", Order: excelgo.PositiveOrder}}
		tqc.Sourc = *sourc
	}
	return writeConfig(&tqc, model.TRIPARTITE_ZHAIPEI_QC_BASE)
}
