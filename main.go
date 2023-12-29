package main

import (
	"PostgresDataHandler/handler"
	"PostgresDataHandler/migrations/initial_migrations"
	"PostgresDataHandler/migrations/post_sync_migrations"
	"database/sql"
	"flag"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
	// Initialize flags and get config values.
	setupFlags()
	pgURI, stateChangeDir, consumerProgressDir, batchBytes, threadLimit, logQueries, readOnlyUserPassword, explorerStatistics, datadogProfiler, isTestnet := getConfigValues()

	// Print all the config values in a single printf call broken up
	// with newlines and make it look pretty both printed out and in code
	glog.Infof(`
		PostgresDataHandler Config Values:
		---------------------------------
		DB_HOST: %s
		DB_PORT: %s
		DB_NAME: %s
		DB_USERNAME: %s
		STATE_CHANGE_DIR: %s
		CONSUMER_PROGRESS_DIR: %s
		BATCH_BYTES: %d
		THREAD_LIMIT: %d
		LOG_QUERIES: %t
		CALCULATE_EXPLORER_STATISTICS: %t
		DATA_DOG_PROFILER: %t
		TESTNET: %t
		`, viper.GetString("DB_HOST"), viper.GetString("DB_PORT"),
		viper.GetString("DB_NAME"), viper.GetString("DB_USERNAME"),
		stateChangeDir, consumerProgressDir, batchBytes, threadLimit,
		logQueries, explorerStatistics, datadogProfiler, isTestnet)

	// Initialize the DB.
	db, err := setupDb(pgURI, threadLimit, logQueries, readOnlyUserPassword, explorerStatistics)
	if err != nil {
		glog.Fatalf("Error setting up DB: %v", err)
	}

	// Setup profiler if enabled.
	if datadogProfiler {
		tracer.Start()
		err := profiler.Start(profiler.WithProfileTypes(profiler.CPUProfile, profiler.BlockProfile, profiler.MutexProfile, profiler.GoroutineProfile, profiler.HeapProfile))
		if err != nil {
			glog.Fatal(err)
		}
	}

	params := &lib.DeSoMainnetParams
	if isTestnet {
		params = &lib.DeSoTestnetParams
	}

	// Initialize and run a state syncer consumer.
	stateSyncerConsumer := &consumer.StateSyncerConsumer{}
	err = stateSyncerConsumer.InitializeAndRun(
		stateChangeDir,
		consumerProgressDir,
		batchBytes,
		threadLimit,
		&handler.PostgresDataHandler{
			DB:     db,
			Params: params,
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

func getConfigValues() (pgURI string, stateChangeDir string, consumerProgressDir string, batchBytes uint64, threadLimit int, logQueries bool, readonlyUserPassword string, explorerStatistics bool, datadogProfiler bool, isTestnet bool) {

	dbHost := viper.GetString("DB_HOST")
	dbPort := viper.GetString("DB_PORT")
	dbName := viper.GetString("DB_NAME")
	dbUsername := viper.GetString("DB_USERNAME")
	dbPassword := viper.GetString("DB_PASSWORD")

	pgURI = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&timeout=18000s", dbUsername, dbPassword, dbHost, dbPort, dbName)

	stateChangeDir = viper.GetString("STATE_CHANGE_DIR")
	if stateChangeDir == "" {
		stateChangeDir = "/tmp/state-changes"
	}

	consumerProgressDir = viper.GetString("CONSUMER_PROGRESS_DIR")
	if consumerProgressDir == "" {
		consumerProgressDir = "/tmp/consumer-progress"
	}

	batchBytes = viper.GetUint64("BATCH_BYTES")
	if batchBytes == 0 {
		batchBytes = 5000000
	}

	threadLimit = viper.GetInt("THREAD_LIMIT")
	if threadLimit == 0 {
		threadLimit = 25
	}

	logQueries = viper.GetBool("LOG_QUERIES")
	readonlyUserPassword = viper.GetString("READONLY_USER_PASSWORD")
	explorerStatistics = viper.GetBool("CALCULATE_EXPLORER_STATISTICS")
	datadogProfiler = viper.GetBool("DATADOG_PROFILER")
	isTestnet = viper.GetBool("IS_TESTNET")

	return pgURI, stateChangeDir, consumerProgressDir, batchBytes, threadLimit, logQueries, readonlyUserPassword, explorerStatistics, datadogProfiler, isTestnet
}

func setupDb(pgURI string, threadLimit int, logQueries bool, readonlyUserPassword string, calculateExplorerStatistics bool) (*bun.DB, error) {
	// Open a PostgreSQL database.
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(pgURI)))
	if pgdb == nil {
		glog.Fatalf("Error connecting to postgres db at URI: %v", pgURI)
	}

	// Create a Bun db on top of postgres for querying.
	db := bun.NewDB(pgdb, pgdialect.New())

	db.SetConnMaxLifetime(0)

	db.SetMaxIdleConns(threadLimit * 2)

	//Print all queries to stdout for debugging.
	if logQueries {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// Set the readonly user password for the initial migrations.
	initial_migrations.SetQueryUserPassword(readonlyUserPassword)

	post_sync_migrations.SetCalculateExplorerStatistics(calculateExplorerStatistics)

	// Apply db migrations.
	err := handler.RunMigrations(db, false, handler.MigrationTypeInitial)
	if err != nil {
		return nil, err
	}
	return db, nil
}
