package database

import "fmt"

type Database interface {
	Start() error
	Stop() error
	Restore([]byte) error
	Cleanup() error
	Query(string) (string, error)
}

type MySQL struct {
	Image string
	Port  string
	User  User
}

type User struct {
	Username string
	Password string
}

func (d MySQL) Start() error {
	return fmt.Errorf("not implemented")
}

func (d MySQL) Stop() error {
	return fmt.Errorf("not implemented")
}

func (d MySQL) Cleanup() error {
	return fmt.Errorf("not implemented")
}

func (d MySQL) Restore(dump []byte) error {
	return fmt.Errorf("not implemented")
}
