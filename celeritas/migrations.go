package celeritas

import (
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (c *Celeritas) MigrateUp(dsn string) error {
	rootPath := filepath.ToSlash(c.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("[migrate.Migrate.Up] => (error): ", err)
		return err
	}

	return nil
}

func (c *Celeritas) MigrateDownAll(dsn string) error {
	rootPath := filepath.ToSlash(c.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err = m.Down(); err != nil {
		log.Println("[migrate.Migrate.Down] => (error): ", err)
		return err
	}

	return nil
}

func (c *Celeritas) Steps(n int, dsn string) error {
	rootPath := filepath.ToSlash(c.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		log.Println("[migrate.Migrate.Steps] => (error): ", err)
		return err
	}

	return nil
}

func (c *Celeritas) MigrateForce(dsn string) error {
	rootPath := filepath.ToSlash(c.RootPath)
	m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		log.Println("[migrate.MigrateForce] => (error): ", err)
		return err
	}

	if err := m.Force(-1); err != nil {
		log.Println("[migrate.Migrate.Force] => (error): ", err)
		return err
	}

	return nil
}
