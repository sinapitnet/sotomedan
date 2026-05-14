package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func GetConfig(fileConfig string) (result *Config) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if fileConfig == "" {
		fileConfig = "main.yml"
	}

	haveExt := false
	if strings.Contains(fileConfig, ".yaml") {
		haveExt = true
	} else if strings.Contains(fileConfig, ".yml") {
		haveExt = true
	}

	if !haveExt {
		fileConfig += ".yaml"
	}

	exPath := filepath.Dir(filepath.Dir(ex))
	fileConfig = filepath.Join(exPath, "config", fileConfig)

	if _, err := os.Stat(fileConfig); err != nil {
		if fileConfig == "" {
			panic("file config not found: " + fileConfig)
		}

		fileConfig = strings.ReplaceAll(fileConfig, ".yaml", ".yml")
	}

	dirCfg := filepath.Dir(fileConfig)
	fileCfg := filepath.Base(fileConfig)
	ext := filepath.Ext(fileCfg)
	fileconfig := strings.TrimSuffix(fileCfg, ext)

	viper.SetConfigName(fileconfig)
	viper.SetConfigType(strings.Replace(ext, ".", "", 1))
	viper.AddConfigPath(dirCfg)

	if err := viper.ReadInConfig(); err != nil {
		panic("unable to read config " + fileConfig)
	}

	if err := viper.Unmarshal(&result); err != nil {
		panic("unable to unmarshal config " + fileConfig)
	}

	return
}
