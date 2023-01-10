package database

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/takutakahashi/database-restore-action/pkg/config"
)

var errDBUndefined error = fmt.Errorf("db is not defined")

type Database struct {
	cfg *config.Config
	db  *sql.DB
}

func genDatabaseURI(cfg *config.Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port)
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
	db, err := sql.Open(string(cfg.Database.Type), genDatabaseURI(cfg))
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
	if err := d.Initialize(); err != nil {
		return err
	}
	if err := d.Restore(); err != nil {
		return err
	}
	defer d.Cleanup()
	return fmt.Errorf("not implemented")
}

func (d Database) Stop() error {
	return fmt.Errorf("not implemented")
}

func (d Database) Initialize() error {
	if d.db == nil {
		return errDBUndefined
	}
	_, err := d.db.Exec(fmt.Sprintf("create database if not exists %s", d.cfg.Database.Name))
	return errors.Wrap(err, "failed to initialize")
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
