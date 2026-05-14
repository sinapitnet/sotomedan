package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveMigration(
	sqls []string,
) (string, string, error) {
	now := time.Now()
	pathVersionYear := now.Year()
	pathVersionMonth := now.Month()
	pathVersionDate := now.Day()
	pathVersion := fmt.Sprintf("%v/%v/%v", pathVersionYear, pathVersionMonth, pathVersionDate)
	version := now.Format(
		"150405",
	)

	filename := fmt.Sprintf(
		"./migrations/%s/%s.sql",
		pathVersion,
		version,
	)

	content := strings.Join(
		sqls,
		"\n\n",
	)

	err := os.MkdirAll(
		filepath.Join(".", "migrations", pathVersion),
		0755,
	)

	if err != nil {
		return "", "", err
	}

	err = os.WriteFile(
		filename,
		[]byte(content),
		0644,
	)

	if err != nil {
		return "", "", err
	}

	return version, filename, nil
}
