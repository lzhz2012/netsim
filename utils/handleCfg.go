package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func LoadConfig(path, cfgFileType string, cfg interface{}) error {
	if cfg == nil {
		logrus.Error("nil pointer")
		return errors.New("Cfg pointer is nil")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("read config file error: %s", err)
		return err
	}
	if strings.ToLower(cfgFileType) == "json" {
		if err := json.Unmarshal(data, cfg); err != nil {
			log.Printf("Load Json Config failed, err:%s", err)
			return err
		}
	} else if strings.ToLower(cfgFileType) == "yml" || strings.ToLower(cfgFileType) == "yaml" {
		if err := yaml.Unmarshal(data, cfg); err != nil { // yml文件有时候读取的时候会存在问题
			log.Printf("Load yaml Config failed, err:%s", err)
			return err
		}
	}

	log.Printf("LoadConfig success")
	return nil
}
