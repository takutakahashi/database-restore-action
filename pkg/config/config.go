package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database Database              `yaml:"database"`
	Check    []DatabaseCheckTarget `yaml:"check"`
	Backup   DatabaseBackup        `yaml:"backup"`
}

type Database struct {
	Type           DatabaseType `yaml:"type"`
	Name           string       `yaml:"name"`
	Image          string       `yaml:"image"`
	User           string       `yaml:"user"`
	Password       string       `yaml:"password"`
	Host           string       `yaml:"host"`
	Port           string       `yaml:"port"`
	DisableRedoLog bool         `yaml:"disable_redo_log"`
}

type DatabaseCheckTarget struct {
	Query    string        `yaml:"query"`
	Operator CheckOperator `yaml:"operator"`
	Value    int           `yaml:"value"`
}

type CheckOperator string

var (
	OpExists CheckOperator = "exists"
	OpEqual  CheckOperator = "equal"
	OpGT     CheckOperator = "gt"
	OpLT     CheckOperator = "lt"
	OpErr    CheckOperator = "error"
	OpNoErr  CheckOperator = "noerror"
)

type DatabaseBackup struct {
	Local string    `yaml:"local,omitempty"`
	S3    BackupS3  `yaml:"s3"`
	Scp   BackupScp `yaml:"scp"`
}

type BackupS3 struct {
	Bucket  string `yaml:"bucket"`
	Key     string `yaml:"key"`
	Profile string `yaml:"profile"`
	Region  string `yaml:"region"`
}

type BackupScp struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
	Path string `yaml:"path"`
	SshKey string `yaml:"sshKey"`
	Key  string `yaml:"key"`
}

type DatabaseType string

var (
	MySQL DatabaseType = "mysql"
)

func Load(p string) (*Config, error) {
	ret := &Config{}
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(buf, ret); err != nil {
		return nil, err
	}
	return ret, nil
}
