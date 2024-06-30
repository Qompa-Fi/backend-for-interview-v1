package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

type DatabaseClient struct {
	db *sql.DB
}

func newDatabaseClient() (DatabaseClient, error) {
	err := os.MkdirAll("./tmp", os.ModePerm)
	if err != nil {
		return DatabaseClient{}, fmt.Errorf("error creating tmp directory: %w", err)
	}

	db, err := sql.Open("sqlite", "./"+filepath.Join("tmp", "database.sqlite"))
	if err != nil {
		return DatabaseClient{}, fmt.Errorf("error opening database: %w", err)
	}

	return DatabaseClient{db}, nil
}

func (c DatabaseClient) Close() error {
	return c.db.Close()
}
