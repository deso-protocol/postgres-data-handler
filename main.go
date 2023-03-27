package main

import (
	"PostgresDataHandler/handler"
	"database/sql"
	"flag"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	// Set glog flags
	flag.Set("log_dir", viper.GetString("log_dir"))
	flag.Set("v", viper.GetString("glog_v"))
	flag.Set("vmodule", viper.GetString("glog_vmodule"))
	flag.Set("alsologtostderr", "true")
	flag.Parse()
	glog.CopyStandardLogTo("INFO")
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	pgURI := viper.GetString("PG_URI")
	if pgURI == "" {
		pgURI = "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	stateChangeFileName := viper.GetString("STATE_CHANGE_FILE_NAME")
	if stateChangeFileName == "" {
		stateChangeFileName = "/tmp/state-changes"
	}

	stateChangeIndexFileName := viper.GetString("STATE_CHANGE_INDEX_FILE_NAME")
	if stateChangeIndexFileName == "" {
		stateChangeIndexFileName = "/tmp/state-changes-index"
	}

	consumerProgressFileName := viper.GetString("CONSUMER_PROGRESS_FILE_NAME")
	if consumerProgressFileName == "" {
		consumerProgressFileName = "/tmp/consumer-progress"
	}

	batchSize := viper.GetInt("BATCH_SIZE")
	if batchSize == 0 {
		batchSize = 10000
	}

	// Open a PostgreSQL database.
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(pgURI)))
	if pgdb == nil {
		glog.Fatalf("Error connecting to postgres db at URI: %v", pgURI)
	}

	// Create a Bun db on top of postgres for querying.
	db := bun.NewDB(pgdb, pgdialect.New())

	// Print all queries to stdout for debugging.
	//db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// Apply db migrations.
	err := handler.RunMigrations(db, false, handler.MigrationTypeInitial)
	if err != nil {
		glog.Fatal(err)
	}

	// Initialize and run a state syncer consumer.
	stateSyncerConsumer := &consumer.StateSyncerConsumer{}
	err = stateSyncerConsumer.InitializeAndRun(
		stateChangeFileName,
		stateChangeIndexFileName,
		consumerProgressFileName,
		true,
		batchSize,
		&handler.PostgresDataHandler{
			DB: db,
		},
	)
	if err != nil {
		glog.Fatal(err)
	}
}
