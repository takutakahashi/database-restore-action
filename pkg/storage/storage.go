package storage

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/takutakahashi/database-restore-action/pkg/config"
)

type S3 struct {
	bucket  string
	key     string
	session *session.Session
}

func NewS3(cfg *config.Config) (*S3, error) {
	if cfg.Backup.S3.Bucket == "" || cfg.Backup.S3.Key == "" {
		return nil, fmt.Errorf("s3 config is not found")
	}
	var opt session.Options
	if cfg.Backup.S3.Profile == "" {
		accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if accessKeyID == "" || secretAccessKey == "" {
			return nil, fmt.Errorf("s3 credential is not found")
		}
		opt = session.Options{
			Config:            aws.Config{Region: aws.String(cfg.Backup.S3.Region)},
			SharedConfigState: session.SharedConfigEnable,
		}
	} else {
		opt = session.Options{
			Config:  aws.Config{Region: aws.String(cfg.Backup.S3.Region)},
			Profile: cfg.Backup.S3.Profile,
		}
	}
	return &S3{
		session: session.Must(session.NewSessionWithOptions(opt)),
		bucket:  cfg.Backup.S3.Bucket,
		key:     cfg.Backup.S3.Key,
	}, nil
}

func (s S3) Download() (string, error) {
	logrus.Infof("downloading file %s/%s", s.bucket, s.key)
	d := s3manager.NewDownloader(s.session)
	f, err := os.CreateTemp("/tmp", "dump")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := d.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	}); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}
