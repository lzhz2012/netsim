package utils

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
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

func Tar(src []string, dst string) error {
	// 创建tar文件
	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	// 通过fw创建一个tar.Writer
	tw := tar.NewWriter(fw)
	// 如果关闭失败会造成tar包不完整
	defer func() {
		if err := tw.Close(); err != nil {
			logrus.Println(err)
		}
	}()

	for _, fileName := range src {
		fi, err := os.Stat(fileName)
		if err != nil {
			logrus.Println(err)
			continue
		}
		hdr, err := tar.FileInfoHeader(fi, "")
		// 将tar的文件信息hdr写入到tw
		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}

		// 将文件数据写入
		fs, err := os.Open(fileName)
		if err != nil {
			return err
		}
		if _, err = io.Copy(tw, fs); err != nil {
			return err
		}
		fs.Close()
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//FIXME: need test
func UnTar(srcFile string, unTarDir string) error {
	//srcFile = "a.tar"
	// open tar package
	fr, err := os.Open(srcFile)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer fr.Close()

	if exist, _ := PathExists(unTarDir); !exist {
		// creat dir
		if err := os.Mkdir(unTarDir, os.ModePerm); err != nil {
			logrus.Error("mkdir failed,", err)
		} else {
			logrus.Printf("mkdir (%s) success!\n", unTarDir)
		}
	}

	PthSep := string(os.PathSeparator)
	tr := tar.NewReader(fr)
	unTarFailFlag := false
	for hdr, err := tr.Next(); err != io.EOF; hdr, err = tr.Next() {
		if err != nil {
			logrus.Error(err)
			unTarFailFlag = true
			continue
		}
		// read file info
		fi := hdr.FileInfo()

		//create a empty file, write unzip data into file
		fw, err := os.Create(unTarDir + PthSep + fi.Name())
		if err != nil {
			logrus.Error(err)
			continue
		}

		if _, err := io.Copy(fw, tr); err != nil {
			logrus.Error(err)
		}
		os.Chmod(fi.Name(), fi.Mode().Perm())
		logrus.Printf("[unZiping-----]unZip to path:%s successful", fw.Name())
		fw.Close()
	}
	if unTarFailFlag {
		logrus.Printf("[unZiping end]unZip file:%s failed", srcFile)
		return fmt.Errorf("UnTar file :%s failed!", srcFile)
	}
	logrus.Printf("[unZiping end]unZip file:%s successful", srcFile)
	return nil
}
