package main

import (
	"fmt"
	"log"

	"github.com/sinapitnet/sotomedan/internal/config"
	"github.com/sinapitnet/sotomedan/internal/db"
	"github.com/sinapitnet/sotomedan/internal/migration"
	"github.com/sinapitnet/sotomedan/internal/schema"
)

func main() {
	configs := config.GetConfig("")
	if configs == nil || configs.Source.Hostname == "" || configs.Destination.Hostname == "" {
		panic("invaid config")
	}

	srcCfg := configs.Source
	dstCfg := configs.Destination

	var srcDsn string
	switch srcCfg.DBEngine {
	case "mysql":
		srcDsn = fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", srcCfg.Username, srcCfg.Password, srcCfg.Hostname, srcCfg.Port, srcCfg.DBName)
	case "sqlserver":
		srcDsn = fmt.Sprintf("sqlserver://%s:%s@%s:%v?database=%s", srcCfg.Username, srcCfg.Password, srcCfg.Hostname, srcCfg.Port, srcCfg.DBName)
	}
	var dstDsn string
	switch dstCfg.DBEngine {
	case "mysql":
		dstDsn = fmt.Sprintf("%s:%s@tcp(%s:%v)/%s", dstCfg.Username, dstCfg.Password, dstCfg.Hostname, dstCfg.Port, dstCfg.DBName)
	case "sqlserver":
		dstDsn = fmt.Sprintf("sqlserver://%s:%s@%s:%v?database=%s", dstCfg.Username, dstCfg.Password, dstCfg.Hostname, dstCfg.Port, dstCfg.DBName)
	}

	cfg := db.Config{
		Source: db.Connection{
			Type: srcCfg.DBEngine,
			DSN:  srcDsn,
		},
		Destination: db.Connection{
			Type: dstCfg.DBEngine,
			DSN:  dstDsn,
		},
	}

	sourceDB, err := db.Open(cfg.Source)
	if err != nil {
		log.Fatal(err)
	}

	destDB, err := db.Open(cfg.Destination)
	if err != nil {
		log.Fatal(err)
	}

	var sourceProvider schema.Provider
	var destProvider schema.Provider

	sourceProvider = schema.NewProvider(cfg.Source.Type)
	destProvider = schema.NewProvider(cfg.Destination.Type)

	sourceSchema, err := sourceProvider.LoadSchema(sourceDB)
	if err != nil {
		log.Fatal(err)
	}

	destSchema, err := destProvider.LoadSchema(destDB)
	if err != nil {
		log.Fatal(err)
	}

	diff := schema.CompareSchemas(
		sourceSchema,
		destSchema,
	)

	sqls := migration.GenerateSQL(
		diff,
		destProvider,
	)

	if len(sqls) == 0 {
		fmt.Println("two database are same version")
		return
	}

	// for _, s := range sqls {
	// 	fmt.Println(s)
	// }

	version, file, err := migration.SaveMigration(sqls)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("migration file:", file)

	err = migration.ExecuteMigration(
		destDB,
		sqls,
		version,
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("migration success")
}
