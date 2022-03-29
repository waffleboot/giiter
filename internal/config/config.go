package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Config struct {
	Log     *os.File
	LogFile string `json:"log"`

	RefreshOnSubject bool
}

func LoadConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	f, err := os.Open(fmt.Sprintf("%s/.giiter/config.json", homeDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&Config); err != nil {
		return err
	}

	if Config.LogFile != "" {
		logFile, err := os.OpenFile(Config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		Config.Log = logFile
	}

	return nil
}

func Close() {
	if Config.Log != nil {
		Config.Log.Close()
	}
}
