package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"rdbak/internal"
)

const configVarName = "CONFIG"
const devConfigPath = "./config.dev.json"

func main() {
	cfgPath := os.Getenv(configVarName)
	if len(cfgPath) == 0 {
		cfgPath = devConfigPath
	}
	cfgJson, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		panic(err)
	}
	var config internal.Config
	if err := json.Unmarshal(cfgJson, &config); err != nil {
		panic(err)
	}

	internal.Backup(&config)
}
