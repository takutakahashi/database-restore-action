package database

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/takutakahashi/database-restore-action/pkg/config"
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
		row := d.db.QueryRow(c.Query)
		count := 0
		err := row.Scan(&count)
		if err2 := pass(c, count, err); err2 != nil {
			return err2
		} else {
			logrus.Info("pass")
		}
	}
	return nil
}

func pass(c config.DatabaseCheckTarget, count int, err error) error {
	logrus.Info(c, count, err)
	if c.Operator != config.OpErr && err != nil {
		return err
	}
	switch c.Operator {
	case config.OpEqual:
		if c.Value != count {
			return fmt.Errorf("expected %d != actual %d", c.Value, count)
		}
		return nil
	case config.OpGT:
		if c.Value > count {
			return fmt.Errorf("expected %d > actual %d", c.Value, count)
		}
		return nil
	case config.OpLT:
		if c.Value < count {
			return fmt.Errorf("expected %d < actual %d", c.Value, count)
		}
		return nil
	case config.OpErr:
		if err == nil {
			return fmt.Errorf("no error is not expected")
		} else {
			logrus.Infof("expected. err: %s", err)
			return nil
		}
	case config.OpNoErr:
		return err
	}
	return fmt.Errorf("not implemented for this op")
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
		command, args := genCommand(d.cfg)
		execCmd := fmt.Sprintf("%s %s < %s", command, strings.Join(args, " "), d.cfg.Backup.Local)
		dumpCmd := exec.Command("bash", "-c", execCmd)
		if buf, err := dumpCmd.Output(); err != nil {
			logrus.Errorf("%s", buf)
			return err
		}
		return nil

	}
	return fmt.Errorf("not implemented")
}
