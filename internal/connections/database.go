package connections

import (
	"database/sql"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func InitDb(config configuration.Config) *sql.DB {
	var err error
	db, err := sql.Open(config.DbDriver, config.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
