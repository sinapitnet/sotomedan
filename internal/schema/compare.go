package schema

type Diff struct {
	CreateTables []Table
	AddColumns   []AddColumn
	ModifyColumn []ModifyColumn
}

type AddColumn struct {
	Table  string
	Column Column
}

type ModifyColumn struct {
	Table  string
	Column Column
}

func CompareSchemas(
	source *DatabaseSchema,
	dest *DatabaseSchema,
) Diff {

	diff := Diff{}

	for tableName, sourceTable := range source.Tables {

		destTable, ok := dest.Tables[tableName]

		if !ok {

			diff.CreateTables = append(
				diff.CreateTables,
				sourceTable,
			)

			continue
		}

		for columnName, sourceColumn := range sourceTable.Columns {

			destColumn, ok := destTable.Columns[columnName]

			if !ok {

				diff.AddColumns = append(
					diff.AddColumns,
					AddColumn{
						Table:  tableName,
						Column: sourceColumn,
					},
				)

				continue
			}

			if needModify(
				sourceColumn,
				destColumn,
			) {

				diff.ModifyColumn = append(
					diff.ModifyColumn,
					ModifyColumn{
						Table:  tableName,
						Column: sourceColumn,
					},
				)
			}
		}
	}

	return diff
}

func needModify(
	a Column,
	b Column,
) bool {

	if a.DataType != b.DataType {
		return true
	}

	if a.IsNullable != b.IsNullable {
		return true
	}

	aDef := ""
	bDef := ""

	if a.DefaultValue != nil {
		aDef = *a.DefaultValue
	}

	if b.DefaultValue != nil {
		bDef = *b.DefaultValue
	}

	return aDef != bDef
}
