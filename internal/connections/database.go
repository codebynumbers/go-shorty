package connections

import (
	"database/sql"
	"github.com/codebynumbers/go-shorty/internal/configuration"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var Db *sql.DB

func InitDb(config configuration.Config) {
	var err error
	Db, err = sql.Open(config.DbDriver, config.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = Db.Ping(); err != nil {
		log.Fatal(err)
	}
}
