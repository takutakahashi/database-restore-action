package storage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	extractedF, err := extract(s.key, f)
	if err != nil {
		return "", err
	}
	return extractedF.Name(), nil
}

func extract(filename string, r *os.File) (*os.File, error) {
	file, ext := splitExt(filename)
	switch ext {
	case ".gz":
		f, err := os.CreateTemp("/tmp", "gz")
		if err != nil {
			return nil, err
		}
		//defer os.Remove(r.Name())

		gr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer gr.Close()

		if _, err := io.Copy(f, gr); err != nil {
			return nil, err
		}
		return extract(file, f)
	case ".tar":
		f, err := os.CreateTemp("/tmp", "tar")
		if err != nil {
			return nil, err
		}
		//defer os.Remove(r.Name())
		tr := tar.NewReader(r)
		if w, err := io.Copy(f, tr); err != nil {
			return nil, err
		} else {
			logrus.Info(w)
		}
		return extract(file, f)
	default:
		return r, nil
	}
}
func splitExt(filename string) (string, string) {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)], ext
}
