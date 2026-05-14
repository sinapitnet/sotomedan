package migration

import "github.com/sinapitnet/sotomedan/internal/schema"

func GenerateSQL(
	diff schema.Diff,
	provider schema.Provider,
) []string {

	var sqls []string

	for _, t := range diff.CreateTables {

		sqls = append(
			sqls,
			provider.GenerateCreateTable(t),
		)
	}

	for _, c := range diff.AddColumns {

		sqls = append(
			sqls,
			provider.GenerateAddColumn(
				c.Table,
				c.Column,
			),
		)
	}

	for _, c := range diff.ModifyColumn {

		sqls = append(
			sqls,
			provider.GenerateModifyColumn(
				c.Table,
				c.Column,
			),
		)
	}

	return sqls
}
