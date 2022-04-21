package app

import (
	"os"

	"gopkg.in/yaml.v2"
)

var Config struct {
	Debug              bool
	Verbose            bool
	EnableGitPush      bool
	UseSubjectToMatch  bool
	MergeRequestPrefix string
	Persistent         struct {
		FeatureBranches []FeatureBranch `yaml:"features"`
	}
}

type FeatureBranch struct {
	BranchName string
	BaseBranch string
}

func LoadConfig(cfgFile string) error {
	f, err := os.Open(cfgFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&Config.Persistent); err != nil {
		return err
	}

	return nil
}
