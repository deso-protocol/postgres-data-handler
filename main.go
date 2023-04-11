package main

import (
	"PostgresDataHandler/handler"
	"database/sql"
	"flag"
	"fmt"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	//"github.com/uptrace/bun/extra/bundebug"
)

func main() {
	// Initialize flags and get config values.
	setupFlags()
	pgURI, stateChangeFileName, stateChangeIndexFileName, consumerProgressFileName, batchSize, threadLimit := getConfigValues()

	// Initialize the DB.
	db, err := setupDb(pgURI, threadLimit)
	if err != nil {
		glog.Fatalf("Error setting up DB: %v", err)
	}

	// Initialize and run a state syncer consumer.
	stateSyncerConsumer := &consumer.StateSyncerConsumer{}
	err = stateSyncerConsumer.InitializeAndRun(
		stateChangeFileName,
		stateChangeIndexFileName,
		consumerProgressFileName,
		batchSize,
		threadLimit,
		&handler.PostgresDataHandler{
			DB: db,
		},
	)
	if err != nil {
		glog.Fatal(err)
	}
}

func setupFlags() {
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
}

func getConfigValues() (pgURI string, stateChangeFileName string, stateChangeIndexFileName string, consumerProgressFileName string, batchSize int, threadLimit int) {
	pgURI = viper.GetString("PG_URI")
	if pgURI == "" {
		pgURI = "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable&timeout=240&connect_timeout=240&write_timeout=240&read_timeout=240&dial_timeout=240"
	}

	stateChangeFileName = viper.GetString("STATE_CHANGE_FILE_NAME")
	if stateChangeFileName == "" {
		stateChangeFileName = "/tmp/state-changes"
	}

	stateChangeIndexFileName = fmt.Sprintf("%s-index", stateChangeFileName)

	consumerProgressFileName = viper.GetString("CONSUMER_PROGRESS_FILE_NAME")
	if consumerProgressFileName == "" {
		consumerProgressFileName = "/tmp/consumer-progress"
	}

	batchSize = viper.GetInt("BATCH_SIZE")
	if batchSize == 0 {
		batchSize = 5000
	}

	threadLimit = viper.GetInt("THREAD_LIMIT")
	if threadLimit == 0 {
		threadLimit = 30
	}
	return pgURI, stateChangeFileName, stateChangeIndexFileName, consumerProgressFileName, batchSize, threadLimit
}

func setupDb(pgURI string, threadLimit int) (*bun.DB, error) {
	// Open a PostgreSQL database.
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(pgURI)))
	if pgdb == nil {
		glog.Fatalf("Error connecting to postgres db at URI: %v", pgURI)
	}

	// Create a Bun db on top of postgres for querying.
	db := bun.NewDB(pgdb, pgdialect.New())

	db.SetConnMaxLifetime(0)

	db.SetMaxIdleConns(threadLimit * 2)

	// Print all queries to stdout for debugging.
	//db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// Apply db migrations.
	err := handler.RunMigrations(db, false, handler.MigrationTypeInitial)
	if err != nil {
		return nil, err
	}
	return db, nil
}
