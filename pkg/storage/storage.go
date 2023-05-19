package storage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	auth "github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"

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

type Scp struct {
	key    string
	host   string
	client scp.Client
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
		logrus.Infof("use profile %s", cfg.Backup.S3.Profile)
		opt = session.Options{
			Config:  aws.Config{Region: aws.String(cfg.Backup.S3.Region)},
			Profile: cfg.Backup.S3.Profile,
		}
	}
	key, err := setKey(cfg.Backup.S3.Key)
	if err != nil {
		return nil, err
	}
	return &S3{
		session: session.Must(session.NewSessionWithOptions(opt)),
		bucket:  cfg.Backup.S3.Bucket,
		key:     key,
	}, nil
}

func setKey(key string) (string, error) {
	parsedKey, err := template.New("tpl").Funcs(template.FuncMap{}).Parse(key)
	if err != nil {
		return "", err
	}
	w := bytes.Buffer{}
	n := time.Now()
	if err := parsedKey.Execute(&w, struct {
		Today       time.Time
		Yesterday   time.Time
		OneWeekAgo  time.Time
		OneMonthAgo time.Time
	}{
		Today:       n,
		Yesterday:   n.AddDate(0, 0, -1),
		OneWeekAgo:  n.AddDate(0, 0, -7),
		OneMonthAgo: n.AddDate(0, -1, 0),
	}); err != nil {
		return "", err
	}
	return w.String(), nil

}

func (s S3) Download() (string, error) {
	logrus.Infof("downloading file %s/%s", s.bucket, s.key)
	d := s3manager.NewDownloader(s.session)
	f, err := os.Create(fmt.Sprintf("/tmp/%s", filepath.Base(s.key)))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if size, err := d.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	}); err != nil {
		os.Remove(f.Name())
		return "", err
	} else {
		logrus.Infof("%d bytes downloaded to %s", size, f.Name())
	}
	return extract(f.Name())
}

func extract(filename string) (string, error) {
	file, ext := splitExt(filename)
	switch ext {
	case ".gz":
		logrus.Infof("extracting %s", filename)
		r, err := os.Open(filename)
		if err != nil {
			return "", err
		}
		f, err := os.Create(file)
		if err != nil {
			return "", err
		}
		defer os.Remove(r.Name())

		gr, err := gzip.NewReader(r)
		if err != nil {
			return "", err
		}
		defer gr.Close()

		if _, err := io.Copy(f, gr); err != nil {
			return "", err
		}
		logrus.Infof("extracted to %s", f.Name())
		return extract(file)
	case ".tar":
		logrus.Infof("extract %s", filename)
		r, err := os.Open(filename)
		if err != nil {
			return "", err
		}
		f, err := os.Create(file)
		if err != nil {
			return "", err
		}
		defer os.Remove(r.Name())
		tr := tar.NewReader(r)
		if _, err := tr.Next(); err == io.EOF {
			return "", err
		}
		if _, err := io.Copy(f, tr); err != nil {
			return "", err
		}
		logrus.Infof("extracted to %s", f.Name())
		return extract(file)
	default:
		return filename, nil
	}
}
func splitExt(filename string) (string, string) {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)], ext
}

func NewScp(cfg *config.Config) (*Scp, error) {
	var clientConfig ssh.ClientConfig
	var err error
	sshKey := os.Getenv("SSH_KEY")
	passPhrase := os.Getenv("SSH_PASSPHRASE")
	password := os.Getenv("SSH_PASSWORD")
	if sshKey != "" {
		if passPhrase != "" {
			clientConfig, err = auth.PrivateKeyWithPassphrase(cfg.Backup.Scp.User, []byte(passPhrase), sshKey, ssh.InsecureIgnoreHostKey())
			if err != nil {
				return nil, err
			}
		} else {
			clientConfig, err = auth.PrivateKey(cfg.Backup.Scp.User, sshKey, ssh.InsecureIgnoreHostKey())
			if err != nil {
				return nil, err
			}
		}
	} else if password != "" {
		clientConfig, err = auth.PrivateKey(cfg.Backup.Scp.User, sshKey, ssh.InsecureIgnoreHostKey())
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no authentication method")
	}

	client := scp.NewClient(fmt.Sprintf("%s:%s", cfg.Backup.Scp.Host, cfg.Backup.Scp.Port), &clientConfig)
	key, err := setKey(cfg.Backup.Scp.Key)
	if err != nil {
		return nil, err
	}

	return &Scp{
		client: client,
		key:    key,
		host:   cfg.Backup.Scp.Host,
	}, nil
}

func (s Scp) Download() (string, error) {
	err := s.client.Connect()
	if err != nil {
		return "", err
	}
	arr1 := strings.Split(s.key, "/")
	fileName := arr1[len(arr1)-1]
	f, err := os.Create(fmt.Sprintf("/tmp/%s", fileName))
	if err != nil {
		return "", err
	}
	defer s.client.Close()
	defer f.Close()
	logrus.Infof("copying %s%s", s.host, s.key)
	err = s.client.CopyFromRemote(context.Background(), f, s.key)
	logrus.Infof("copyid %s%s", s.host, s.key)
	if err != nil {
		return "", err
	}

	return extract(f.Name())
}
