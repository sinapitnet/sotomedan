package utils

import (
	"fmt"
	"strings"
)

func NormalizeMysqlValue(value string, dataType string) (newValue string) {
	newValue = fmt.Sprintf("'%s'", value)
	val := strings.ToUpper(value)
	if strings.Contains(val, "NULL") ||
		strings.Contains(val, "CURRENT_TIMESTAMP") {
		newValue = value
	}

	if strings.Contains(strings.ToLower(dataType), "bit") ||
		strings.Contains(strings.ToLower(dataType), "enum") {
		newValue = value
	}

	return newValue
}

func NormalizeMySQLType(
	dataType string,
	length int64,
	precision int64,
	scale int64,
) string {
	t := strings.ToLower(strings.TrimSpace(dataType))

	switch t {

	// =========================
	// STRING
	// =========================

	case "varchar":

		// Hindari row size overflow
		if length >= 500 {
			return "TEXT"
		}

		return fmt.Sprintf("varchar(%d)", length)

	case "char":
		return fmt.Sprintf("char(%d)", length)

	case "text", "tinytext", "mediumtext", "longtext":
		return t

	// =========================
	// INTEGER
	// =========================

	case "tinyint":

		// tinyint(1) biasanya boolean
		if length == 1 {
			return "tinyint(1)"
		}

		return "tinyint"

	case "smallint":
		return "smallint"

	case "mediumint":
		return "mediumint"

	case "int", "integer":
		return "int"

	case "bigint":
		return "bigint"

	// =========================
	// DECIMAL
	// =========================

	case "decimal", "numeric":

		if precision > 0 {
			return fmt.Sprintf(
				"decimal(%d,%d)",
				precision,
				scale,
			)
		}

		return "decimal"

	// =========================
	// FLOATING
	// =========================

	case "float":
		return "float"

	case "double":
		return "double"

	// =========================
	// DATE TIME
	// =========================

	case "date":
		return "date"

	case "datetime":
		return "datetime"

	case "timestamp":
		return "timestamp"

	case "time":
		return "time"

	case "year":
		return "year"

	// =========================
	// BOOLEAN / BIT
	// =========================

	case "bit":

		if length > 0 {
			return fmt.Sprintf("bit(%d)", length)
		}

		return "bit(1)"

	case "boolean", "bool":
		return "tinyint(1)"

	// =========================
	// JSON
	// =========================

	case "json":
		return "json"

	// =========================
	// BINARY
	// =========================

	case "blob", "tinyblob", "mediumblob", "longblob":
		return t

	case "binary":
		return fmt.Sprintf("binary(%d)", length)

	case "varbinary":
		return fmt.Sprintf("varbinary(%d)", length)

	// =========================
	// UUID
	// =========================

	case "uuid":
		return "char(36)"

	// =========================
	// FALLBACK
	// =========================

	default:
		return t
	}
}
