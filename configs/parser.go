package configs

import (
	"apiring/utils"
	"fmt"
	"github.com/Unknwon/goconfig"
	"os"
	"path/filepath"
)


func ParserConfig() *goconfig.ConfigFile {
	dir := utils.AbsPath()
	ini := filepath.Join(dir, "config.ini")
	config, err := goconfig.LoadConfigFile(ini)
	if err != nil {
		fmt.Println("Get config file error")
		os.Exit(-1)
	}
	return config
}
