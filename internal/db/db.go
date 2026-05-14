package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/microsoft/go-mssqldb"
)

func Open(c Connection) (*sql.DB, error) {
	var driver string

	switch c.Type {
	case "mysql":
		driver = "mysql"
	case "sqlserver":
		driver = "sqlserver"
	default:
		return nil, fmt.Errorf("unsupported db type")
	}

	db, err := sql.Open(driver, c.DSN)
	if err != nil {
		return nil, err
	}

	return db, nil
}
