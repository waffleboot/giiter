package app

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var Config AppConfig

type AppConfig struct {
	Repo          string
	Push          bool
	Debug         bool
	Verbose       bool
	BaseBranch    string
	FeatureBranch string
	Prefix        string
}

func LoadConfig(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.WithMessage(err, "get user home dir")
		}

		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName(".giiter")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return errors.WithMessage(err, "read config")
	}

	if err := viper.UnmarshalExact(&Config); err != nil {
		return errors.WithMessage(err, "parse config file")
	}

	return nil
}
