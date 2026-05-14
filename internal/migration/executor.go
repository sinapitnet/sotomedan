package migration

import (
	"database/sql"
	"fmt"
	"time"
)

func ExecuteMigration(
	db *sql.DB,
	sqls []string,
	version string,
) error {

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, _ = tx.Exec(`
CREATE TABLE IF NOT EXISTS schema_versions (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	version VARCHAR(100),
	executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
`)

	n := 0
	for i, q := range sqls {
		query := fmt.Sprintf("Execute query: %s\n\n", q)
		fmt.Println(query)

		_, err := tx.Exec(q)

		if err != nil {
			return fmt.Errorf(
				"migration failed at query %d: %w",
				n,
				err,
			)
		}

		if i%10 == 0 && i != 0 {
			n = i + 1
			fmt.Printf("%v. sleep progress......\n", n)
			time.Sleep(10 * time.Second)
		}
	}

	_, err = tx.Exec(
		`
INSERT INTO schema_versions(version)
VALUES(?)
`,
		version,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
