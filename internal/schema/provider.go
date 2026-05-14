package schema

import (
	"database/sql"
	"fmt"
)

type Provider interface {
	LoadTablesDef(db *sql.DB) ([]TableDef, error)
	LoadTableDef(tableName string) TableDef
	LoadTables(db *sql.DB) ([]string, error)
	LoadSchema(db *sql.DB) (*DatabaseSchema, error)

	GenerateCreateTable(table Table) string

	GenerateAddColumn(
		table string,
		column Column,
	) string

	GenerateModifyColumn(
		table string,
		column Column,
	) string
}

func NewProvider(dbType string) Provider {
	switch dbType {
	case "mysql":
		return &MySQLProvider{}
	case "sqlserver":
		return &SQLServerProvider{}
	}

	panic(fmt.Sprintf("unsupported provider: %s", dbType))
}
