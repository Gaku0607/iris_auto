package routers

import (
	"os"
	"path/filepath"

	"github.com/Gaku0607/iris_auto/model"
	"github.com/Gaku0607/iris_auto/tool"
)

func MakeServiceRouters() error {

	//shipp-list
	if err := MakeServiceRouter(model.SHIPP_LIST_MOTHOD); err != nil {
		return err
	}

	//spilt-and-export
	if err := MakeServiceRouter(model.SPILT_AND_EXPORT_MOTHOD); err != nil {
		return err
	}

	//wenda-qc-table
	if err := MakeServiceRouter(model.WENDA_QC_MOTHOD); err != nil {
		return err
	}

	//zhaipei-qc-table
	if err := MakeServiceRouter(model.ZHAIPEI_QC_MOTHOD); err != nil {
		return err
	}

	//tripartite-form
	if err := MakeServiceRouter(model.TRIPARTITE_FORM_MOTHOD); err != nil {
		return err
	}

	if err := MakeServiceRouter(model.THIRD_PARTY_QC_MOTHOD); err != nil {
		return err
	}

	//返回檔案地址
	if exist := tool.IsExist(model.Result_Path); !exist {
		if err := os.MkdirAll(model.Result_Path, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func MakeServiceRouter(method string) error {
	path := absoluteServicePath(method)
	if exist := tool.IsExist(path); !exist {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func absoluteServicePath(method string) string {
	return filepath.Join(model.Services_Dir, method)
}
