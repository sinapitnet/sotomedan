package schema

type Column struct {
	Name         string
	DataType     string
	IsNullable   bool
	DefaultValue *string
	IsPrimaryKey bool
	IsKey        bool
	ExtraCol     string
}

type Table struct {
	Name    string
	Columns map[string]Column
	Engine  string
	Collate string
	Charset string
}

type DatabaseSchema struct {
	Tables map[string]Table
}

type TableDef struct {
	TableName string
	Engine    string
	Collate   string
	Charset   string
}
