package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

type Config struct {
	Token string `json:"token"`
}

var (
	configDir string
)

func init() {
	dir, e := os.UserHomeDir()
	if e == nil {
		configDir = path.Join(dir, ".fconfig")
	}
}

func ReadConfig() Config {
	var config Config
	fRead, _ := os.Open(configDir)
	bytes, _ := ioutil.ReadAll(fRead)
	json.Unmarshal(bytes, &config)
	return config
}

func (f *Config) WriteConfig() {
	c, e := json.Marshal(f)
	if e != nil {
		fil, err := os.Create(configDir)
		if err != nil {
			fil.Write(c)
			fil.Close()
		}
	}
}
