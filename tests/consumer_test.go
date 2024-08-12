package tests

import (
	"fmt"
	"github.com/deso-protocol/backend/config"
	"github.com/deso-protocol/backend/routes"
	coreCmd "github.com/deso-protocol/core/cmd"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/postgres-data-handler/entries"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
	"math"
	"math/rand"
	"testing"
	"time"
)

const (
	globalStateSharedSecret = "abcdef"
)

type TestHandler struct {
	// Params is a struct containing the current blockchain parameters.
	// It is used to determine which prefix to use for public keys.
	Params                *lib.DeSoParams
	HandleEntryBatchEmit  func(batchedEntries []*lib.StateChangeEntry, emittedIdx int) error
	handleEntryBatchCount int
}

// TODO: Figure out how to inject some tests into this (probably through the testHandler struct).
func (th *TestHandler) HandleEntryBatch(batchedEntries []*lib.StateChangeEntry) error {
	err := th.HandleEntryBatchEmit(batchedEntries, th.handleEntryBatchCount)
	if err != nil {
		return err
	}
	th.handleEntryBatchCount++

	uniqueEntries := consumer.UniqueEntries(batchedEntries)
	// Create a new array to hold the bun struct.

	// Loop through the entries and convert them to PGPostEntry.
	for _, entry := range uniqueEntries {

		if entry.OperationType == lib.DbOperationTypeDelete {
			switch entry.EncoderType {
			case lib.EncoderTypeProfileEntry:
				fmt.Printf("Deleting profile Entry: %+v\n", entry.KeyBytes)
			case lib.EncoderTypeBalanceEntry:
				fmt.Printf("Deleting balance Entry: %+v\n", entry.KeyBytes)
			}
			continue
		}

		switch entry.EncoderType {
		case lib.EncoderTypeProfileEntry:
			pgEntry := entries.ProfileEntryEncoderToPGStruct(entry.Encoder.(*lib.ProfileEntry), entry.KeyBytes, th.Params)
			fmt.Printf("Profile Entry: %+v\n", pgEntry)
			fmt.Printf("Entry: %+v\n", entry)
			if entry.AncestralRecord == nil {
				fmt.Printf("Ancestral record is nil\n")
				fmt.Printf("Ancestral record bytes: %+v\n", entry.AncestralRecordBytes)
			} else {
				ancestralRecordEntry := entries.ProfileEntryEncoderToPGStruct(entry.AncestralRecord.(*lib.ProfileEntry), entry.KeyBytes, th.Params)
				fmt.Printf("Ancestral record: %+v\n", ancestralRecordEntry)
			}
		case lib.EncoderTypeBalanceEntry:
			pgEntry := entries.BalanceEntryEncoderToPGStruct(entry.Encoder.(*lib.BalanceEntry), entry.KeyBytes, th.Params)
			fmt.Printf("Balance: %+v\n", pgEntry)
		}

	}

	return nil
}

func (th *TestHandler) HandleSyncEvent(syncEvent consumer.SyncEvent) error {
	fmt.Printf("Handling sync event: %v\n", syncEvent)
	return nil
}

func (th *TestHandler) InitiateTransaction() error {
	//fmt.Printf("Initiating transaction\n")
	return nil
}

func (th *TestHandler) CommitTransaction() error {
	//fmt.Printf("Committing transaction\n")
	return nil
}

func (th *TestHandler) RollbackTransaction() error {
	//fmt.Printf("Rolling back transaction\n")
	return nil
}

func (th *TestHandler) GetParams() *lib.DeSoParams {
	return th.Params
}

func RandString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestConsumer(t *testing.T) {
	// TODO: Figure out how to clean all this up - make it more in-tune with the rest of the test utils.
	desoParams := &lib.DeSoTestnetParams
	starterSeed := "verb find card ship another until version devote guilt strong lemon six"
	seedBytes, err := bip39.NewSeedWithErrorChecking(starterSeed, "")
	require.NoError(t, err)
	publicKey, _, _, err := ComputeKeysFromSeedWithNet(seedBytes, 0, desoParams)
	require.NoError(t, err)

	publicKeyBase58 := lib.Base58CheckEncode(publicKey.SerializeCompressed(), false, desoParams)
	require.NoError(t, err)

	stateDirPostFix := RandString(10)

	stateChangeDir := fmt.Sprintf("./ss/state-changes-%s-%s", t.Name(), stateDirPostFix)
	consumerProgressDir := fmt.Sprintf("./ss/consumer-progress-%s-%s", t.Name(), stateDirPostFix)

	apiServer, _ := newTestApiServer(t, publicKeyBase58, 17001, stateChangeDir)
	require.NoError(t, err)

	// Start the api server in a non-blocking way.
	go func() {
		apiServer.Start()
	}()

	//stateChangeSyncer := nodeServer.StateChangeSyncer

	testConfig, err := SetupTestEnvironment(3, fmt.Sprintf("%s%d", t.Name(), rand.Intn(math.MaxInt16)), false)
	require.NoError(t, err)

	nodeClient := testConfig.NodeClient
	coinUser := testConfig.TestUsers[0]

	// Mint some DAO coins for the coin user.
	mintDaoCoinReq := &routes.DAOCoinRequest{
		UpdaterPublicKeyBase58Check:           coinUser.PublicKeyBase58,
		ProfilePublicKeyBase58CheckOrUsername: coinUser.PublicKeyBase58,
		OperationType:                         routes.DAOCoinOperationStringMint,
		CoinsToMintNanos:                      *uint256.NewInt().SetUint64(1e12),
		TransferRestrictionStatus:             routes.TransferRestrictionStatusStringUnrestricted,
		MinFeeRateNanosPerKB:                  FeeRateNanosPerKB,
		TransactionFees:                       nil,
	}

	_, _, err = nodeClient.DAOCoins(mintDaoCoinReq, coinUser.PrivateKey, false, true)
	require.NoError(t, err)

	// TODO: Create a locked balance for the test users.

	// Lock coins for the coin user.
	lockCoinsReq := &routes.CoinLockupRequest{
		TransactorPublicKeyBase58Check: coinUser.PublicKeyBase58,
		ProfilePublicKeyBase58Check:    coinUser.PublicKeyBase58,
		RecipientPublicKeyBase58Check:  coinUser.PublicKeyBase58,
		UnlockTimestampNanoSecs:        time.Now().UnixNano() + 1000000000000,
		VestingEndTimestampNanoSecs:    time.Now().UnixNano() + 1000000000000,
		LockupAmountBaseUnits:          uint256.NewInt().SetUint64(1e9),
		ExtraData:                      nil,
		MinFeeRateNanosPerKB:           FeeRateNanosPerKB,
		TransactionFees:                nil,
	}

	_, _, err = nodeClient.LockCoins(lockCoinsReq, coinUser.PrivateKey, false, true)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	testHandlerFunc := func(batchedEntries []*lib.StateChangeEntry, emittedIdx int) error {
		fmt.Printf("Handling %d batched entries of type %d for idx: %d\n", len(batchedEntries), batchedEntries[0].EncoderType, emittedIdx)

		return nil
	}

	testHandler := &TestHandler{
		Params:                desoParams,
		handleEntryBatchCount: 0,
		HandleEntryBatchEmit:  testHandlerFunc,
	}

	stateSyncerConsumer := &consumer.StateSyncerConsumer{}
	err = stateSyncerConsumer.InitializeAndRun(
		stateChangeDir,
		consumerProgressDir,
		500000,
		1,
		testHandler,
	)
	require.NoError(t, err)

	time.Sleep(2 * time.Hour)
}

func newTestApiServer(t *testing.T, starterUserPublicKeyBase58 string, apiPort uint16, stateChangeDir string) (*routes.APIServer, *lib.Server) {
	// Create a badger db instance.
	badgerDB, badgerDir := routes.GetTestBadgerDb(t)

	// Set core node's config.
	coreConfig := coreCmd.LoadConfig()
	coreConfig.Params = &lib.DeSoTestnetParams
	coreConfig.DataDirectory = badgerDir
	coreConfig.Regtest = true
	coreConfig.TXIndex = false
	coreConfig.DisableNetworking = true
	coreConfig.MinerPublicKeys = []string{starterUserPublicKeyBase58}
	coreConfig.NumMiningThreads = 1
	coreConfig.HyperSync = false
	coreConfig.MinFeerate = 2000
	coreConfig.LogDirectory = badgerDir
	coreConfig.StateChangeDir = stateChangeDir

	// Create a core node.
	shutdownListener := make(chan struct{})
	node := coreCmd.NewNode(coreConfig)
	node.Start(&shutdownListener)

	// Set api server's config.
	apiConfig := config.LoadConfig(coreConfig)
	apiConfig.APIPort = apiPort
	apiConfig.GlobalStateRemoteNode = ""
	apiConfig.GlobalStateRemoteSecret = globalStateSharedSecret
	apiConfig.RunHotFeedRoutine = false
	apiConfig.RunSupplyMonitoringRoutine = false
	apiConfig.AdminPublicKeys = []string{starterUserPublicKeyBase58}
	apiConfig.SuperAdminPublicKeys = []string{starterUserPublicKeyBase58}

	// Create an api server.
	apiServer, err := routes.NewAPIServer(
		node.Server,
		node.Server.GetMempool(),
		node.Server.GetBlockchain(),
		node.Server.GetBlockProducer(),
		node.TXIndex,
		node.Params,
		apiConfig,
		node.Config.MinFeerate,
		badgerDB,
		nil,
		node.Config.BlockCypherAPIKey,
	)
	require.NoError(t, err)

	// Initialize api server.
	apiServer.MinFeeRateNanosPerKB = node.Config.MinFeerate

	t.Cleanup(func() {
		apiServer.Stop()
		node.Stop()
	})
	return apiServer, node.Server
}
