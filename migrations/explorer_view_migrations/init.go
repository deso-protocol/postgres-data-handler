package explorer_view_migrations

import (
	"github.com/uptrace/bun/migrate"
)

var Migrations = migrate.NewMigrations()

var (
	DbHost     string
	DbPort     string
	DbUsername string
	DbPassword string
	DbName     string
)

func SetDBConfig(host string, port string, username string, password string, dbname string) {
	DbHost = host
	DbPort = port
	DbUsername = username
	DbPassword = password
	DbName = dbname
}

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}
