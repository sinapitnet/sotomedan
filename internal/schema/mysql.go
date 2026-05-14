package schema

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/sinapitnet/sotomedan/internal/utils"
)

type MySQLProvider struct {
	tablesDef []TableDef
}

func (s *MySQLProvider) LoadTablesDef(db *sql.DB) ([]TableDef, error) {
	var result []TableDef
	var err error

	query := `
SELECT
    t.TABLE_NAME,
    t.ENGINE,
    t.TABLE_COLLATION as 'COLLATE',
    c.CHARACTER_SET_NAME as 'CHARSET'
FROM information_schema.TABLES t
JOIN information_schema.COLLATION_CHARACTER_SET_APPLICABILITY c
    ON t.TABLE_COLLATION = c.COLLATION_NAME
WHERE t.TABLE_SCHEMA = Database()
AND t.TABLE_TYPE = 'BASE TABLE'
ORDER BY t.TABLE_NAME 
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, engine, collate, charset string

		err := rows.Scan(
			&tableName,
			&engine,
			&collate,
			&charset,
		)

		if err != nil {
			return result, err
		}

		result = append(result, TableDef{
			TableName: tableName,
			Engine:    engine,
			Collate:   collate,
			Charset:   charset,
		})
	}

	return result, err
}

func (s *MySQLProvider) LoadTableDef(tableName string) TableDef {
	var result TableDef

	for _, t := range s.tablesDef {
		if t.TableName == tableName {
			return t
		}
	}

	return result
}

func (s *MySQLProvider) LoadTables(db *sql.DB) ([]string, error) {
	var result []string
	var err error

	query := `
SELECT TABLE_NAME
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE()
AND TABLE_TYPE = 'BASE TABLE'
ORDER BY TABLE_NAME
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(
			&tableName,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, tableName)
	}

	return result, err
}

func (s *MySQLProvider) LoadSchema(
	db *sql.DB,
) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Tables: map[string]Table{},
	}

	tables, err := s.LoadTables(db)
	if len(tables) == 0 {
		return schema, err
	}
	tablesDef, err := s.LoadTablesDef(db)
	if err != nil {
		return schema, err
	}
	s.tablesDef = tablesDef

	for _, t := range tables {
		fmt.Printf("Load Schema Table %s\n", t)
		tablesDef := s.LoadTableDef(t)
		query := `
SELECT
	c.TABLE_NAME,
	c.COLUMN_NAME,
	c.COLUMN_TYPE,
	c.IS_NULLABLE,
	c.COLUMN_DEFAULT,
	c.COLUMN_KEY,
	c.EXTRA 
FROM INFORMATION_SCHEMA.COLUMNS c
WHERE c.TABLE_SCHEMA = DATABASE()
AND c.TABLE_NAME = '` + t + `'
ORDER BY c.TABLE_NAME, c.ORDINAL_POSITION
`

		rows, err := db.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			var columnName string
			var dataType string
			var isNullable string
			var defaultValue sql.NullString
			var columnKey string
			var extraCol string

			err := rows.Scan(
				&tableName,
				&columnName,
				&dataType,
				&isNullable,
				&defaultValue,
				&columnKey,
				&extraCol,
			)

			if err != nil {
				return nil, err
			}

			table, ok := schema.Tables[tableName]

			if !ok {

				table = Table{
					Name:    tableName,
					Columns: map[string]Column{},
					Engine:  tablesDef.Engine,
					Collate: tablesDef.Collate,
					Charset: tablesDef.Charset,
				}
			}

			var def *string

			if defaultValue.Valid {
				v := defaultValue.String
				def = &v
			}

			table.Columns[columnName] = Column{
				Name:         columnName,
				DataType:     dataType,
				IsNullable:   isNullable == "YES",
				DefaultValue: def,
				IsPrimaryKey: columnKey == "PRI",
				IsKey:        columnKey == "MUL",
				ExtraCol:     extraCol,
			}

			schema.Tables[tableName] = table
		}
	}

	return schema, nil
}

func (s *MySQLProvider) GenerateCreateTable(
	table Table,
) string {
	var extraKey string
	var cols []string
	var priCols []string

	for _, c := range table.Columns {
		line := fmt.Sprintf(
			"\t`%s` %s",
			c.Name,
			c.DataType,
		)

		if !c.IsNullable {
			line += " NOT NULL"
		}

		if c.DefaultValue != nil {
			defaultVal := utils.NormalizeMysqlValue(*c.DefaultValue, c.DataType)

			line += fmt.Sprintf(
				" DEFAULT %s",
				defaultVal,
			)
		}

		if c.IsPrimaryKey {
			extraKey = c.Name
		}

		if c.IsKey {
			priCols = append(priCols, c.Name)
		}

		if c.ExtraCol != "" {
			line += fmt.Sprintf(
				" %s",
				c.ExtraCol,
			)

			if c.ExtraCol == "auto_increment" && extraKey == "" {
				extraKey = c.Name
			}
		}

		cols = append(cols, line)
	}

	// if table.Name == "tprtproductionorderreworksectionexecuted" {
	// 	log.Println("tprtproductionorderreworksectionexecuted")
	// }

	if extraKey != "" {
		if len(priCols) != 0 {
			idx := slices.Index(priCols, extraKey)
			// remove primary key
			if idx != -1 {
				priCols = append(priCols[:idx], priCols[idx+1:]...)
			}
		}

		// always set primary key to index 0
		priCols = append([]string{extraKey}, priCols...)
		cols = append(cols, fmt.Sprintf("\tPRIMARY KEY (`%s`)", strings.Join(priCols, "`,`")))
	}

	var def []string
	if table.Engine != "" {
		def = append(def, fmt.Sprintf("ENGINE=%s", table.Engine))
	}
	if table.Charset != "" {
		def = append(def, fmt.Sprintf("CHARSET=%s", table.Charset))
	}
	if table.Collate != "" {
		def = append(def, fmt.Sprintf("COLLATE=%s", table.Collate))
	}

	return fmt.Sprintf(
		"CREATE TABLE `%s` (\n%s\n) %s;",
		table.Name,
		strings.Join(cols, ",\n"),
		strings.Join(def, " "),
	)
}

func (s *MySQLProvider) GenerateAddColumn(
	table string,
	column Column,
) string {

	sql := fmt.Sprintf(
		"ALTER TABLE `%s` ADD COLUMN `%s` %s",
		table,
		column.Name,
		column.DataType,
	)

	if !column.IsNullable {
		sql += " NOT NULL"
	}

	if column.DefaultValue != nil {
		defaultVal := utils.NormalizeMysqlValue(*column.DefaultValue, column.DataType)
		sql += fmt.Sprintf(
			" DEFAULT %s",
			defaultVal,
		)
	}

	sql += ";"

	return sql
}

func (s *MySQLProvider) GenerateModifyColumn(
	table string,
	column Column,
) string {

	sql := fmt.Sprintf(
		"ALTER TABLE `%s` MODIFY COLUMN `%s` %s",
		table,
		column.Name,
		column.DataType,
	)

	if !column.IsNullable {
		sql += " NOT NULL"
	}

	if column.DefaultValue != nil {
		defaultVal := utils.NormalizeMysqlValue(*column.DefaultValue, column.DataType)
		sql += fmt.Sprintf(
			" DEFAULT %s",
			defaultVal,
		)
	}

	sql += ";"

	return sql
}
