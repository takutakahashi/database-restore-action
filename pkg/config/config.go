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
	Type     DatabaseType `yaml:"type"`
	Name     string       `yaml:"name"`
	Image    string       `yaml:"image"`
	User     string       `yaml:"user"`
	Password string       `yaml:"password"`
	Host     string       `yaml:"host"`
	Port     string       `yaml:"port"`
}

type DatabaseCheckTarget struct {
	Database string `yaml:"database"`
	Table    string `yaml:"table"`
	Column   string `yaml:"column"`
	CountGT  int    `yaml:"count_gt"`
	CountLT  int    `yaml:"count_lt"`
	CountEQ  int    `yaml:"count_eq"`
}

type DatabaseBackup struct {
	Local        string       `yaml:"local"`
	URI          string       `yaml:"uri"`
	S3Credential S3Credential `yaml:"s3_credential"`
}

type S3Credential struct {
	AccessKeyId     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Profile         string `yaml:"profile"`
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
