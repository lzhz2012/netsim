package utils

import (
	"errors"
	"io/ioutil"
	"os"
)

//获取指定目录下的所有文件和目录(支持递归遍历，但是要考虑遍历深度，可能会出现因为文件夹嵌套太深堆栈不够用)
func GetFilesAndDirs(dirPth string, files *[]string, dirs *[]string) error {
	if files == nil || dirs == nil {
		return errors.New("nil pointer")
	}
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			*dirs = append(*dirs, dirPth+PthSep+fi.Name())
			GetFilesAndDirs(dirPth+PthSep+fi.Name(), files, dirs)
		} else { //fi.IsRegular()
			// 过滤指定格式
			// ok := strings.HasSuffix(fi.Name(), ".go")
			// if ok {
			// 	*files = append(*files, dirPth+PthSep+fi.Name())
			// }
			*files = append(*files, dirPth+PthSep+fi.Name())
		}
	}

	return nil
}
