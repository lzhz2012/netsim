package utils

import (
	"fmt"
	"log"
	"testing"
)

func TestGetFilesAndDirs(t *testing.T) {
	var files, dirs []string

	transerveDir := "D:\\code\\go_test"
	transerveDir = "."
	if err := GetFilesAndDirs(transerveDir, &files, &dirs); err != nil {
		fmt.Printf("遍历文件夹失败，错误原因：%s", err)
	}
	fmt.Printf("获取的文件夹为[%s]\n", dirs)
	fmt.Printf("获取的文件为[%s]\n", files)

}

func TestTarAndUnTar(t *testing.T) {

	var files, dirs []string

	transerveDir := "D:\\code\\go_test"
	transerveDir = "."
	if err := GetFilesAndDirs(transerveDir, &files, &dirs); err != nil {
		fmt.Printf("遍历文件夹失败，错误原因：%s", err)
	}

	dst := "a.tar"
	if err := Tar(files, dst); err != nil {
		log.Fatal(err)
	}

	unTarPath := "D:\\code\\netsim\\utils\\unTar"
	if err := UnTar(dst, unTarPath); err != nil {
		log.Fatal(err)
	}
}
