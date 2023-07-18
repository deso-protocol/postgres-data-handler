package initial_migrations

import "github.com/uptrace/bun/migrate"

var (
	queryUserPassword string
	Migrations        = migrate.NewMigrations()
)

func SetQueryUserPassword(password string) {
	queryUserPassword = password
}

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
}
