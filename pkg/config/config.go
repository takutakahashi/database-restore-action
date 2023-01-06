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
	Type  DatabaseType `yaml:"type"`
	Image string       `yaml:"image"`
}

type DatabaseCheckTarget struct {
	Table   string `yaml:"table"`
	Column  string `yaml:"column"`
	CountGT int    `yaml:"count_gt"`
	CountLT int    `yaml:"count_lt"`
	CountEQ int    `yaml:"count_eq"`
}

type DatabaseBackup struct {
	URI          string       `json:"uri"`
	S3Credential S3Credential `json:"s3_credential"`
}

type S3Credential struct {
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
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
