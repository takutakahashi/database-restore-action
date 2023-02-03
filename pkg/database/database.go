package database

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/takutakahashi/database-restore-action/pkg/config"
	"github.com/takutakahashi/database-restore-action/pkg/storage"
)

var errDBUndefined error = fmt.Errorf("db is not defined")

type Database struct {
	cfg *config.Config
	db  *sql.DB
}

func genDatabaseURI(cfg *config.Config, includeDatabaseName bool) string {
	dbname := ""
	if includeDatabaseName {
		dbname = cfg.Database.Name
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, dbname)
}

func genCommand(cfg *config.Config) (string, []string) {
	if cfg.Database.Type == config.MySQL {
		return "mysql", []string{
			"-h",
			cfg.Database.Host,
			"-u",
			cfg.Database.User,
			fmt.Sprintf("-p%s", cfg.Database.Password),
			"-P",
			cfg.Database.Port,
			cfg.Database.Name,
		}
	}
	return "", nil
}
func New(cfg *config.Config) (Database, error) {
	d := Database{
		cfg: cfg,
	}
	db, err := sql.Open(string(cfg.Database.Type), genDatabaseURI(cfg, true))
	if err != nil {
		return d, err
	}
	d.db = db
	return d, nil
}

func (d Database) Start() error {
	return fmt.Errorf("not implemented")
}

func (d Database) Run() error {
	defer d.Cleanup()
	if err := d.Initialize(); err != nil {
		return err
	}
	if err := d.Restore(); err != nil {
		return err
	}
	if err := d.RunTest(); err != nil {
		return err
	}
	return nil
}

func (d Database) Stop() error {
	return fmt.Errorf("not implemented")
}

func (d Database) Initialize() error {
	if d.db == nil {
		return errDBUndefined
	}
	if err := d.db.Ping(); err != nil {
		tmpDB, err := sql.Open(string(d.cfg.Database.Type), genDatabaseURI(d.cfg, false))
		if err != nil {
			return err
		}
		if _, err := tmpDB.Exec(fmt.Sprintf("create database if not exists %s", d.cfg.Database.Name)); err != nil {
			return err
		}
	}
	return nil
}

func (d Database) RunTest() error {
	for _, c := range d.cfg.Check {
		row, err := d.db.Query(c.Query)
		if err != nil {
			return err
		}
		count := 0
		for row.Next() {
			count++
		}
		if err2 := pass(c, count, err); err2 != nil {
			return err2
		}
	}
	return nil
}

func pass(c config.DatabaseCheckTarget, count int, err error) error {
	if c.Operator != config.OpErr && err != nil {
		return err
	}
	switch c.Operator {
	case config.OpExists:
		if count <= 0 {
			return fmt.Errorf("no row exists")
		} else {
			logrus.Infof("[pass] row exists, count: %d, query: %s", count, c.Query)
			return nil
		}
	case config.OpEqual:
		if c.Value != count {
			return fmt.Errorf("expected %d != actual %d", c.Value, count)
		}
	case config.OpGT:
		if c.Value > count {
			return fmt.Errorf("expected %d > actual %d", c.Value, count)
		}
	case config.OpLT:
		if c.Value < count {
			return fmt.Errorf("expected %d < actual %d", c.Value, count)
		}
	case config.OpErr:
		if err == nil {
			return fmt.Errorf("no error is not expected")
		} else {
			return nil
		}
	case config.OpNoErr:
		return err
	default:
		return fmt.Errorf("not implemented for this op")
	}
	logrus.Infof("[pass] %s = %d, %s %d", c.Query, count, c.Operator, c.Value)
	return nil
}

func (d Database) Cleanup() error {
	if d.db == nil {
		return errDBUndefined
	}
	_, err := d.db.Exec(fmt.Sprintf("drop database %s", d.cfg.Database.Name))
	return err
}

func (d Database) Restore() error {
	if d.cfg.Backup.Local != "" {
		return restoreLocal(d.cfg, d.cfg.Backup.Local)

	}
	if d.cfg.Backup.S3.Bucket != "" && d.cfg.Backup.S3.Key != "" {
		s, err := storage.NewS3(d.cfg)
		if err != nil {
			return err
		}
		localPath, err := s.Download()
		if err != nil {
			return err
		}
		defer os.Remove(localPath)
		return restoreLocal(d.cfg, localPath)

	}
	return fmt.Errorf("not implemented")
}

func restoreLocal(cfg *config.Config, path string) error {
	command, args := genCommand(cfg)
	execCmd := fmt.Sprintf("%s %s < %s", command, strings.Join(args, " "), path)
	logrus.Info("execting restore...")
	dumpCmd := exec.Command("bash", "-c", execCmd)
	if buf, err := dumpCmd.CombinedOutput(); err != nil {
		logrus.Errorf("restore failed. error: %s", err)
		logrus.Errorf("%s", buf)
		return err
	}
	logrus.Info("restored")
	return nil
}
