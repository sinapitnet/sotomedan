package schema

import (
	"database/sql"
	"fmt"
	"strings"
)

type SQLServerProvider struct {
}

func (s *SQLServerProvider) LoadTablesDef(db *sql.DB) ([]TableDef, error) {
	var result []TableDef
	var err error

	return result, err
}

func (s *SQLServerProvider) LoadTableDef(tableName string) TableDef {
	var result TableDef

	return result
}

func (s *SQLServerProvider) LoadTables(db *sql.DB) ([]string, error) {
	var result []string
	var err error

	return result, err
}

func (s *SQLServerProvider) LoadSchema(
	db *sql.DB,
) (*DatabaseSchema, error) {

	query := `
SELECT
	c.TABLE_NAME,
	c.COLUMN_NAME,
	c.DATA_TYPE +
	CASE
		WHEN c.CHARACTER_MAXIMUM_LENGTH IS NOT NULL
			THEN '(' +
				CASE
					WHEN c.CHARACTER_MAXIMUM_LENGTH = -1
						THEN 'MAX'
					ELSE CAST(c.CHARACTER_MAXIMUM_LENGTH AS VARCHAR)
				END + ')'
		WHEN c.NUMERIC_PRECISION IS NOT NULL
			THEN '(' + CAST(c.NUMERIC_PRECISION AS VARCHAR) +
				CASE
					WHEN c.NUMERIC_SCALE IS NOT NULL
						THEN ',' + CAST(c.NUMERIC_SCALE AS VARCHAR)
					ELSE ''
				END + ')'
		ELSE ''
	END AS COLUMN_TYPE,
	c.IS_NULLABLE,
	c.COLUMN_DEFAULT,
	CASE
		WHEN pk.COLUMN_NAME IS NOT NULL THEN 'PRI'
		ELSE ''
	END AS COLUMN_KEY,
	COLUMNPROPERTY(
		OBJECT_ID(c.TABLE_SCHEMA + '.' + c.TABLE_NAME),
		c.COLUMN_NAME,
		'IsIdentity'
	) AS IS_IDENTITY
FROM INFORMATION_SCHEMA.COLUMNS c
LEFT JOIN (
	SELECT ku.TABLE_NAME, ku.COLUMN_NAME
	FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
	JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
		ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
	WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
) pk
	ON c.TABLE_NAME = pk.TABLE_NAME
	AND c.COLUMN_NAME = pk.COLUMN_NAME
ORDER BY c.TABLE_NAME, c.ORDINAL_POSITION
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schema := &DatabaseSchema{
		Tables: map[string]Table{},
	}

	for rows.Next() {
		var tableName string
		var columnName string
		var dataType string
		var isNullable string
		var defaultValue sql.NullString
		var columnKey string
		var isIdentity int

		err := rows.Scan(
			&tableName,
			&columnName,
			&dataType,
			&isNullable,
			&defaultValue,
			&columnKey,
			&isIdentity,
		)

		if err != nil {
			return nil, err
		}

		table, ok := schema.Tables[tableName]

		if !ok {
			table = Table{
				Name:    tableName,
				Columns: map[string]Column{},
			}
		}

		var def *string

		if defaultValue.Valid {
			v := defaultValue.String
			def = &v
		}

		isPrimaryKey := columnKey == "PRI"

		extra := ""
		if isIdentity == 1 {
			extra = "IDENTITY(1,1)"
		}

		table.Columns[columnName] = Column{
			Name:         columnName,
			DataType:     dataType,
			IsNullable:   isNullable == "YES",
			DefaultValue: def,
			IsPrimaryKey: isPrimaryKey,
			ExtraCol:     extra,
		}

		schema.Tables[tableName] = table
	}

	return schema, nil
}

func (s *SQLServerProvider) GenerateCreateTable(
	table Table,
) string {
	var cols []string
	var priCols []string

	for _, c := range table.Columns {

		line := fmt.Sprintf(
			"\t[%s] %s",
			c.Name,
			c.DataType,
		)

		if c.ExtraCol != "" {
			line += " " + c.ExtraCol
		}

		if !c.IsNullable {
			line += " NOT NULL"
		}

		if c.DefaultValue != nil {
			line += fmt.Sprintf(
				" DEFAULT %s",
				*c.DefaultValue,
			)
		}

		if c.IsPrimaryKey {
			priCols = append(
				priCols,
				fmt.Sprintf("[%s]", c.Name),
			)
		}

		cols = append(cols, line)
	}

	if len(priCols) != 0 {
		cols = append(
			cols,
			fmt.Sprintf(
				"\tPRIMARY KEY (%s)",
				strings.Join(priCols, ","),
			),
		)
	}

	return fmt.Sprintf(
		"CREATE TABLE [%s] (\n%s\n);",
		table.Name,
		strings.Join(cols, ",\n"),
	)
}

func (s *SQLServerProvider) GenerateAddColumn(
	table string,
	column Column,
) string {
	sql := fmt.Sprintf(
		"ALTER TABLE [%s] ADD [%s] %s",
		table,
		column.Name,
		column.DataType,
	)

	if column.ExtraCol != "" {
		sql += " " + column.ExtraCol
	}

	if !column.IsNullable {
		sql += " NOT NULL"
	}

	if column.DefaultValue != nil {
		sql += fmt.Sprintf(
			" DEFAULT %s",
			*column.DefaultValue,
		)
	}

	sql += ";"

	return sql
}

func (s *SQLServerProvider) GenerateModifyColumn(
	table string,
	column Column,
) string {
	sql := fmt.Sprintf(
		"ALTER TABLE [%s] ALTER COLUMN [%s] %s",
		table,
		column.Name,
		column.DataType,
	)

	if !column.IsNullable {
		sql += " NOT NULL"
	} else {
		sql += " NULL"
	}

	sql += ";"

	return sql
}
