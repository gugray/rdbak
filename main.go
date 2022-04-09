package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"rdbak/internal"
)

const configVarName = "CONFIG"
const devConfigPath = "./config.dev.json"
const pwdKeyHex = "566D597133743677397A244226452948404D635166546A576E5A723475377821"

var cfg internal.Config

type task func()

func main() {
	action := getActionFromArgs()
	readConfig()
	action()
}

func getActionFromArgs() task {
	if len(os.Args) != 2 {
		return usage
	}
	switch os.Args[1] {
	case "backup":
		return backup
	case "encrypt-pwd":
		return encryptPwd
	default:
		return usage
	}
}

func usage() {
	fmt.Printf(`RDBAK is a tool to create a local backup of your Raindrop bookmarks and permanent copies.

Usage:

	rdbak <command>

RDBAK will read the config file specified in the %v environment variable. 
If the variable is not spefied, it will look for config.dev.json in its working directory.

The commands are:

	backup			Retrieves changes to your bookmarks and downloads new permanent copies
	encrypt-pwd		Replaces your plaintext password in the config file with an encrypted string

`,
		configVarName)
}

func encryptPwd() {

	if len(cfg.Password) == 0 {
		panic("Config has no plaintext password")
	}
	if len(cfg.EncryptedPassword) != 0 {
		panic("Config already has an encrypted password")
	}

	cfg.EncryptedPassword = internal.Encrypt(cfg.Password, pwdKeyHex)
	cfg.Password = ""

	cfgPath := os.Getenv(configVarName)
	if len(cfgPath) == 0 {
		cfgPath = devConfigPath
	}
	cfgJson, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(cfgPath, cfgJson, 0666)
	if err != nil {
		panic(err)
	}
}

func readConfig() {

	cfgPath := os.Getenv(configVarName)
	if len(cfgPath) == 0 {
		cfgPath = devConfigPath
	}
	cfgJson, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(cfgJson, &cfg); err != nil {
		panic(err)
	}

	if len(cfg.Password) == 0 {
		cfg.Password = internal.Decrypt(cfg.EncryptedPassword, pwdKeyHex)
	}
}

func backup() {
	internal.Backup(&cfg)
}
