package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strings"

	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/postgres-data-handler/handler"
	"github.com/deso-protocol/postgres-data-handler/migrations/initial_migrations"
	"github.com/deso-protocol/postgres-data-handler/migrations/post_sync_migrations"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/golang/glog"
	lru "github.com/hashicorp/golang-lru/v2"
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
	pgURI, stateChangeDir, consumerProgressDir, batchBytes, threadLimit, logQueries, readOnlyUserPassword,
		explorerStatistics, datadogProfiler, isTestnet, isRegtest, isAcceleratedRegtest, syncMempool := getConfigValues()

	// Print all the config values in a single printf call broken up
	// with newlines and make it look pretty both printed out and in code
	glog.Infof(`
		PostgresDataHandler Config Values:
		---------------------------------
		DB_HOST: %s
		DB_PORT: %s
		DB_USERNAME: %s
		STATE_CHANGE_DIR: %s
		CONSUMER_PROGRESS_DIR: %s
		BATCH_BYTES: %d
		THREAD_LIMIT: %d
		LOG_QUERIES: %t
		CALCULATE_EXPLORER_STATISTICS: %t
		DATA_DOG_PROFILER: %t
		TESTNET: %t
		REGTEST: %t
		ACCELERATED_REGTEST: %t
		`, viper.GetString("DB_HOST"), viper.GetString("DB_PORT"),
		viper.GetString("DB_USERNAME"),
		stateChangeDir, consumerProgressDir, batchBytes, threadLimit,
		logQueries, explorerStatistics, datadogProfiler, isTestnet, isRegtest, isAcceleratedRegtest)

	// Initialize the DB.
	db, err := setupDb(pgURI, threadLimit, logQueries, readOnlyUserPassword, explorerStatistics)
	if err != nil {
		glog.Fatalf("Error setting up DB: %v", err)
	}

	// Setup profiler if enabled.
	if datadogProfiler {
		tracer.Start()
		err = profiler.Start(profiler.WithProfileTypes(profiler.CPUProfile, profiler.BlockProfile, profiler.MutexProfile, profiler.GoroutineProfile, profiler.HeapProfile))
		if err != nil {
			glog.Fatal(err)
		}
	}

	params := &lib.DeSoMainnetParams
	if isTestnet {
		params = &lib.DeSoTestnetParams
		if isRegtest {
			params.EnableRegtest(isAcceleratedRegtest)
		}
	}
	lib.GlobalDeSoParams = *params

	cachedEntries, err := lru.New[string, []byte](int(handler.EntryCacheSize))
	if err != nil {
		glog.Fatalf("Error creating LRU cache: %v", err)
	}

	// Initialize and run a state syncer consumer.
	stateSyncerConsumer := &consumer.StateSyncerConsumer{}
	err = stateSyncerConsumer.InitializeAndRun(
		stateChangeDir,
		consumerProgressDir,
		batchBytes,
		threadLimit,
		syncMempool,
		&handler.PostgresDataHandler{
			DB:            db,
			Params:        params,
			CachedEntries: cachedEntries,
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

func getConfigValues() (pgURI string, stateChangeDir string, consumerProgressDir string, batchBytes uint64, threadLimit int, logQueries bool, readonlyUserPassword string, explorerStatistics bool, datadogProfiler bool, isTestnet bool, isRegtest bool, isAcceleratedRegtest bool, syncMempool bool) {

	dbHost := viper.GetString("DB_HOST")
	dbPort := viper.GetString("DB_PORT")
	dbUsername := viper.GetString("DB_USERNAME")
	dbPassword := viper.GetString("DB_PASSWORD")

	pgURI = fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable&timeout=18000s", dbUsername, dbPassword, dbHost, dbPort)

	stateChangeDir = viper.GetString("STATE_CHANGE_DIR")
	if stateChangeDir == "" {
		stateChangeDir = "/tmp/state-changes"
	}
	// Set the state change dir flag that core uses, so DeSoEncoders properly encode and decode state change metadata.
	viper.Set("state-change-dir", stateChangeDir)

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

	syncMempool = viper.GetBool("SYNC_MEMPOOL")

	logQueries = viper.GetBool("LOG_QUERIES")
	readonlyUserPassword = viper.GetString("READONLY_USER_PASSWORD")
	explorerStatistics = viper.GetBool("CALCULATE_EXPLORER_STATISTICS")
	datadogProfiler = viper.GetBool("DATADOG_PROFILER")
	isTestnet = viper.GetBool("IS_TESTNET")
	isRegtest = viper.GetBool("REGTEST")
	isAcceleratedRegtest = viper.GetBool("ACCELERATED_REGTEST")

	return pgURI, stateChangeDir, consumerProgressDir, batchBytes, threadLimit, logQueries, readonlyUserPassword, explorerStatistics, datadogProfiler, isTestnet, isRegtest, isAcceleratedRegtest, syncMempool
}

type CustomQueryHook struct {
	bundebug.QueryHook
}

func (h *CustomQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	queryStr := event.Query

	if strings.HasPrefix(strings.ToUpper(queryStr), "BEGIN") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "COMMIT") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "SELECT PG_ADVISORY") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "SAVEPOINT") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "RELEASE SAVEPOINT") {
		return ctx
	}
	return h.QueryHook.BeforeQuery(ctx, event)
}

func (h *CustomQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	queryStr := event.Query
	if strings.HasPrefix(strings.ToUpper(queryStr), "BEGIN") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "COMMIT") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "SELECT PG_ADVISORY") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "SAVEPOINT") ||
		strings.HasPrefix(strings.ToUpper(queryStr), "RELEASE SAVEPOINT") {
		return
	}
	h.QueryHook.AfterQuery(ctx, event)
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
		// customHook := &CustomQueryHook{
		// 	QueryHook: *bundebug.NewQueryHook(bundebug.WithVerbose(true)),
		// }
		// db.AddQueryHook(customHook)
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
