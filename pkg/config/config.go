package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DatabaseType  DatabaseType          `json:"database_type"`
	DatabaseImage string                `json:"database_image"`
	CheckTargets  []DatabaseCheckTarget `json:"check_targets"`
	Backup        DatabaseBackup        `json:"backup"`
}

type DatabaseCheckTarget struct {
	Table   string `json:"table"`
	Column  string `json:"column"`
	CountGT int    `json:"count_gt"`
	CountLT int    `json:"count_lt"`
	CountEQ int    `json:"count_eq"`
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
