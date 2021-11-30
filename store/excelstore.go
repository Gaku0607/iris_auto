package store

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Gaku0607/excelgo"
	iris "github.com/Gaku0607/iris_auto"
	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/shipplist"
	"github.com/xuri/excelize/v2"
)

//全局儲存變量
var Store = &ExcelStore{}

type ExcelStore struct {
}

func NewExcelStore() *ExcelStore {
	e := &ExcelStore{}
	return e
}

func (e *ExcelStore) addCell(sheet, addr string, val interface{}, to *excelize.File) error {
	return to.SetCellValue(sheet, addr, val)
}

//匯出指定的出倉單格式
func (e *ExcelStore) ExportShippList(foramtfile *excelize.File, shippDetail *shipplist.ShippingDetail, Tcol []string, path string) error {

	sheetid := 0
	//每筆出倉單最大件數
	const maxheigh = 20

	for area, wh := range shippDetail.Areas {

		if iris.NEXT2 == area {
			sheetid = 0
		} else {
			sheetid = 1
		}

		if err := wh.Format(); err != nil {
			return err
		}

		data := wh.Rows

		for i := 0; ; i++ {
			if len(data) <= maxheigh {
				if err := e.exportshipplist(foramtfile, wh, Tcol, sheetid, data); err != nil {
					return err
				}
				break
			} else {
				if err := e.exportshipplist(foramtfile, wh, Tcol, sheetid, data[:maxheigh]); err != nil {
					return err
				}
				data = data[maxheigh:]
			}
		}
	}

	foramtfile.DeleteSheet(foramtfile.GetSheetName(0))
	foramtfile.DeleteSheet(foramtfile.GetSheetName(0))

	if err := foramtfile.SaveAs(path); err != nil {
		return err
	}

	return nil
}

func (e *ExcelStore) exportshipplist(f *excelize.File, area *shipplist.Warehouse, Tcol []string, sheetid int, data [][]interface{}) error {

	sl := model.Environment.SL.ShippListPosition
	sheetname := area.Area

	//查看是否有為重複Sheet
	for i := 1; ; i++ {

		if idx := f.GetSheetIndex(sheetname); idx != -1 {
			sheetname = area.Area + "(" + strconv.Itoa(i) + ")"
		} else {
			break
		}

	}

	newsheetid := e.CreateSheet(f, sheetname)
	if err := f.CopySheet(sheetid, newsheetid); err != nil {
		return err
	}

	//輸入細節
	//收件人
	rec, _ := f.GetCellValue(sheetname, sl.RecipientPosition)
	if err := e.addCell(sheetname, sl.RecipientPosition, rec+area.Recipient, f); err != nil {
		return err
	}
	//地址
	addr, _ := f.GetCellValue(sheetname, sl.AddrPosition)
	if err := e.addCell(sheetname, sl.AddrPosition, addr+area.Addr, f); err != nil {
		return err
	}
	//收件人電話
	tel, _ := f.GetCellValue(sheetname, sl.TelPosition)
	if err := e.addCell(sheetname, sl.TelPosition, tel+area.Tel, f); err != nil {
		return err
	}
	//總數
	if err := e.addCell(sheetname, sl.TotalPosition, area.Total(data), f); err != nil {
		return err
	}
	//出倉單號
	if err := e.addCell(sheetname, sl.OrderNumPosition, area.OrderNumber(), f); err != nil {
		return err
	}
	//出倉日期
	if err := e.addCell(sheetname, sl.DatePosition, area.Date(), f); err != nil {
		return err
	}
	//負責人
	if err := e.addCell(sheetname, sl.PartyPosition, "System", f); err != nil {
		return err
	}

	for i, row := range data {
		for index, val := range row {
			if Tcol[index] != "" {
				if err := e.addCell(sheetname, e.getaddr(Tcol[index], sl.DataPosition+i), val, f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//匯出穩達ＱＣ檔案
func (e *ExcelStore) ExportWendaFiles(foramtpath, path string, s *excelgo.Sourc, data [][]interface{}) error {
	//每頁的最大列數
	const WENDA_DRIVER_FILE_HIGH = 45

	//回傳給穩達時所使用的文件格式
	file, err := excelize.OpenFile(foramtpath)
	if err != nil {
		return err
	}

	sheet := file.GetSheetName(0)

	addcount := len(data) - WENDA_DRIVER_FILE_HIGH
	if addcount > 0 {
		name := file.GetDefinedName()[file.GetSheetIndex(model.WENDA_DRIVER_SHEET)]
		index := strings.LastIndex(name.RefersTo, "$")
		if index != -1 {
			high, err := strconv.Atoi(name.RefersTo[index+1:])
			if err != nil {
				return err
			}

			name.RefersTo = name.RefersTo[:index+1] + strconv.Itoa(high+addcount)
			if err := e.modityDefinedName(file, &name); err != nil {
				return err
			}
			if err := e.insertRows(addcount, 2, file, model.WENDA_DRIVER_SHEET); err != nil {
				return err
			}
		}
	}

	wendaparms := model.Environment.IDS.WendaMergeBox
	//匯出回傳穩達的檔案
	for i, row := range data {
		i += 2
		//客代 CustomerCode （固定值）
		if err := e.addCell(sheet, e.getaddr(wendaparms.CustomerCodeTCol, i), wendaparms.CustomerCode, file); err != nil {
			return err
		}
		//是否自取 （固定值）
		if err := e.addCell(sheet, e.getaddr(wendaparms.IsSelfCollectionTCol, i), wendaparms.IsSelfCollection, file); err != nil {
			return err
		}
		//宅配安裝 （固定值）
		if err := e.addCell(sheet, e.getaddr(wendaparms.DeliveryInstallationTCol, i), wendaparms.DeliveryInstallation, file); err != nil {
			return err
		}

		tcolfn := s.IteratorByTCol()
		for {
			tcol, exist := tcolfn()
			if !exist {
				break
			}

			if err := e.addCell(tcol.Sheet, e.getaddr(tcol.TColStr, i), tcol.Format(row[tcol.ParentCol.Col]), file); err != nil {
				return err
			}
		}

		for _, f := range s.Formulas {
			if err := e.addCellFormula(file, f.TSheet, e.getaddr(f.TColStr, i), f.Formulafn(i, f.FormulaStr)); err != nil {
				return err
			}
		}
	}

	//設置總數的BOXTOTAL公式
	if err := e.addCellFormula(
		file,
		model.WENDA_DRIVER_SHEET,
		e.getaddr(s.Formulas[0].TColStr, addcount+WENDA_DRIVER_FILE_HIGH+2),
		fmt.Sprintf(model.Environment.IDS.BoxTotalFormula, 2, addcount+WENDA_DRIVER_FILE_HIGH+1),
	); err != nil {
		return err
	}

	sheetId := file.NewSheet(model.WENDA_QC_SHEET)
	//匯出ＱＣ
	if err := e.exportExcel(file, sheetId, s.SetHeader(data)); err != nil {
		return err
	}
	//變更Sheet排序
	file.WorkBook.Sheets.Sheet[0], file.WorkBook.Sheets.Sheet[sheetId] = file.WorkBook.Sheets.Sheet[sheetId], file.WorkBook.Sheets.Sheet[0]
	return file.SaveAs(path)
}

// //匯出宅配通ＱＣ檔案
func (e *ExcelStore) ExportZhaipeiFiles(path string, s *excelgo.Sourc, QCdata [][]interface{}) error {

	file := excelize.NewFile()
	file.SetSheetName(model.ORIGIN_SHEET, model.ZHAIPEI_QC_SHEET)
	//匯出ＱＣ
	sheetid := file.GetSheetIndex(model.ZHAIPEI_QC_SHEET)
	if err := e.exportExcel(file, sheetid, s.SetHeader(QCdata)); err != nil {
		return err
	}

	//匯出回傳宅配通格式
	sheetid = file.NewSheet(model.ZHAIPEI_RETURNS_SHEET)

	for i, row := range QCdata {
		i += 1
		tcolfn := s.IteratorByTCol()
		for {
			tcol, exist := tcolfn()
			if !exist {
				break
			}

			if err := e.addCell(tcol.Sheet, e.getaddr(tcol.TColStr, i), tcol.Format(row[tcol.ParentCol.Col]), file); err != nil {
				return err
			}
		}

		for _, f := range s.Formulas {
			if err := e.addCellFormula(file, f.TSheet, e.getaddr(f.TColStr, i), f.Formulafn(i+1, f.FormulaStr)); err != nil {
				return err
			}
		}
	}

	return file.SaveAs(path)
}

//對內容進行全部匯出
func (e *ExcelStore) Export(path string, data [][]interface{}) error {
	nf := excelize.NewFile()
	Id := nf.NewSheet(model.ORIGIN_SHEET)
	if err := e.exportExcel(nf, Id, data); err != nil {
		return err
	}
	return nf.SaveAs(path)
}

//已行的方式寫入檔案
func (e *ExcelStore) addrow(f *excelize.File, sheetid int, addr string, row []interface{}) {
	f.SetSheetRow(f.GetSheetName(sheetid), addr, &row)
}

//已Excel的格式匯出
func (e *ExcelStore) exportExcel(nf *excelize.File, sheetId int, rows [][]interface{}) error {
	for i, row := range rows {
		e.addrow(nf, sheetId, "A"+strconv.Itoa(i+1), row)
	}
	return nil
}

//設置公式
func (e *ExcelStore) addCellFormula(file *excelize.File, sheet, addr, formula string) error {
	return file.SetCellFormula(sheet, addr, formula)
}

//修改Sheet的定義名稱
func (e *ExcelStore) modityDefinedName(file *excelize.File, name *excelize.DefinedName) error {

	if err := file.DeleteDefinedName(&excelize.DefinedName{Scope: name.Scope, Name: name.Name}); err != nil {
		return err
	}
	return file.SetDefinedName(name)
}

//插入列
func (e *ExcelStore) insertRows(count, Position int, file *excelize.File, sheet string) error {
	for i := 0; i < count; i++ {
		if err := file.DuplicateRow(sheet, Position); err != nil {
			return err
		}
	}
	return nil
}

func (e *ExcelStore) CreateSheet(file *excelize.File, name string) int {
	return e.createSheet(file, name, name, 0)
}

func (e *ExcelStore) createSheet(file *excelize.File, name string, base string, id int) int {
	sheetid := 0
	if sheetid = file.NewSheet(name); sheetid == 0 {
		return e.createSheet(file, fmt.Sprintf("%s(%d)", base, id), base, id+1)
	}
	return sheetid
}

func (e *ExcelStore) getaddr(header string, index int) string {
	return header + strconv.Itoa(index)
}
