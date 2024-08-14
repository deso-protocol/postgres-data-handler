package tests

import (
	"fmt"
	"github.com/deso-protocol/backend/config"
	"github.com/deso-protocol/backend/routes"
	coreCmd "github.com/deso-protocol/core/cmd"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/postgres-data-handler/entries"
	"github.com/deso-protocol/state-consumer/consumer"
	"github.com/google/uuid"
	"github.com/holiman/uint256"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
	"math"
	"math/rand"
	"testing"
)

const (
	globalStateSharedSecret = "abcdef"
)

type TestHandler struct {
	// Params is a struct containing the current blockchain parameters.
	// It is used to determine which prefix to use for public keys.
	Params *lib.DeSoParams

	BatchedEntryChan chan *HandleEntryBatchEvent
	SyncEventChan    chan consumer.SyncEvent
	InitiateTxnChan  chan struct{}
	CommitTxnChan    chan struct{}
	RollbackTxnChan  chan struct{}
}

func NewTestHandler(params *lib.DeSoParams) *TestHandler {
	th := &TestHandler{}
	th.Params = params

	th.BatchedEntryChan = make(chan *HandleEntryBatchEvent)
	th.SyncEventChan = make(chan consumer.SyncEvent)
	th.InitiateTxnChan = make(chan struct{})
	th.CommitTxnChan = make(chan struct{})
	th.RollbackTxnChan = make(chan struct{})
	return th
}

type HandleEntryBatchEvent struct {
	BatchedEntries []*lib.StateChangeEntry
	IsMempool      bool
}

func (th *TestHandler) HandleEntryBatch(batchedEntries []*lib.StateChangeEntry, isMempool bool) error {
	// Add the batched entries to the channel.
	th.BatchedEntryChan <- &HandleEntryBatchEvent{
		BatchedEntries: batchedEntries,
		IsMempool:      isMempool,
	}

	return nil
}

func (th *TestHandler) HandleSyncEvent(syncEvent consumer.SyncEvent) error {
	//th.SyncEventChan <- syncEvent
	return nil
}

func (th *TestHandler) InitiateTransaction() error {
	//th.InitiateTxnChan <- struct{}{}
	return nil
}

func (th *TestHandler) CommitTransaction() error {
	//th.CommitTxnChan <- struct{}{}
	return nil
}

func (th *TestHandler) RollbackTransaction() error {
	//th.RollbackTxnChan <- struct{}{}
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

	_, txnRes, err := nodeClient.DAOCoins(mintDaoCoinReq, coinUser.PrivateKey, false, true)
	require.NoError(t, err)

	// TODO: Create a locked balance for the test users.

	// Lock coins for the coin user.
	//lockCoinsReq := &routes.CoinLockupRequest{
	//	TransactorPublicKeyBase58Check: coinUser.PublicKeyBase58,
	//	ProfilePublicKeyBase58Check:    coinUser.PublicKeyBase58,
	//	RecipientPublicKeyBase58Check:  coinUser.PublicKeyBase58,
	//	UnlockTimestampNanoSecs:        time.Now().UnixNano() + 1000000000000,
	//	VestingEndTimestampNanoSecs:    time.Now().UnixNano() + 1000000000000,
	//	LockupAmountBaseUnits:          uint256.NewInt().SetUint64(1e9),
	//	ExtraData:                      nil,
	//	MinFeeRateNanosPerKB:           FeeRateNanosPerKB,
	//	TransactionFees:                nil,
	//}
	//
	//_, txnRes, err := nodeClient.LockCoins(lockCoinsReq, coinUser.PrivateKey, false, true)
	//require.NoError(t, err)

	txnHash := txnRes.TxnHashHex

	fmt.Printf("Looking for txn: %s\n", txnHash)

	testHandler := NewTestHandler(desoParams)

	stateSyncerConsumer := &consumer.StateSyncerConsumer{}

	// Initialize and run the state syncer consumer in a non-blocking thread.
	go func() {
		err := stateSyncerConsumer.InitializeAndRun(
			stateChangeDir,
			consumerProgressDir,
			500000,
			1,
			testHandler,
		)
		require.NoError(t, err)
	}()

	//targetEncoderType := lib.EncoderTypeBalanceEntry
	//nextBalanceEntryEvent, err := testHandler.WaitForMatchingEntryBatch(&targetEncoderType, nil, nil)
	//require.NoError(t, err)
	//fmt.Printf("BALANCE ENTRY EVENT: %d|%d|%t|%t|%s\n", nextBalanceEntryEvent.EncoderType, nextBalanceEntryEvent.OperationType, nextBalanceEntryEvent.IsMempool, nextBalanceEntryEvent.EntryBatch[0].IsReverted, nextBalanceEntryEvent.FlushId.String())
	//
	//balanceEntry := lib.EncoderTypeBalanceEntry.New()
	//
	//// Convert byte slice to bytes.Reader.
	//err = consumer.DecodeEntry(balanceEntry, nextBalanceEntryEvent.EntryBatch[0].EncoderBytes)
	//require.NoError(t, err)
	//fmt.Printf("BALANCE ENTRY: %+v\n", balanceEntry)
	//
	//for ii := 0; ii < 100000000; ii++ {
	//	nextBalanceEntryEvent, err = testHandler.WaitForMatchingEntryBatch(&targetEncoderType, nil, nil)
	//	require.NoError(t, err)
	//	fmt.Printf("BALANCE ENTRY EVENT: %d|%d|%t|%t|%s\n", nextBalanceEntryEvent.EncoderType, nextBalanceEntryEvent.OperationType, nextBalanceEntryEvent.IsMempool, nextBalanceEntryEvent.EntryBatch[0].IsReverted, nextBalanceEntryEvent.FlushId.String())
	//	// Convert byte slice to bytes.Reader.
	//	err = consumer.DecodeEntry(balanceEntry, nextBalanceEntryEvent.EntryBatch[0].EncoderBytes)
	//	require.NoError(t, err)
	//	fmt.Printf("BALANCE ENTRY: %+v\n", balanceEntry)
	//}

	coinLockTxnRes, err := testHandler.WaitForTxnHash(txnHash)
	require.NoError(t, err)
	fmt.Printf("FOUND THE TRANSACTION: %+v\n", coinLockTxnRes.Txn)

	for _, entry := range coinLockTxnRes.EntryBatch {
		fmt.Printf("Entry: %+v\n", entry)
	}

	for _, remainingTxn := range coinLockTxnRes.RemainingTxns {
		fmt.Printf("REMAINING TRANSACTION: %+v\n", remainingTxn)
	}

	nextBatches := testHandler.GetNextBatches(50)
	for _, nextBatch := range nextBatches {
		if nextBatch.EncoderType == lib.EncoderTypeBalanceEntry {

		}
		fmt.Printf("NEXT ENTRY BATCH: %d|%d|%t|%s\n", nextBatch.EncoderType, nextBatch.OperationType, nextBatch.IsMempool, nextBatch.FlushId.String())
	}
}

// EntryScanResult is a struct that contains the results of a search for a transaction or entry in the BatchedEntryChan.
type EntryScanResult struct {
	BatchesScanned int
	EntryBatch     []*lib.StateChangeEntry
	IsMempool      bool
	EncoderType    lib.EncoderType
	OperationType  lib.StateSyncerOperationType
	Txn            *entries.PGTransactionEntry
	RemainingTxns  []*entries.PGTransactionEntry
	FlushId        uuid.UUID
}

// GetNextBatch returns the next batch of entries from the BatchedEntryChan.
func (th *TestHandler) GetNextBatch() *EntryScanResult {
	batchEvent := <-th.BatchedEntryChan
	batchedEntries := batchEvent.BatchedEntries
	return &EntryScanResult{
		EntryBatch:     batchedEntries,
		IsMempool:      batchEvent.IsMempool,
		BatchesScanned: 1,
		FlushId:        batchedEntries[0].FlushId,
		EncoderType:    batchedEntries[0].EncoderType,
		OperationType:  batchedEntries[0].OperationType,
	}
}

func (th *TestHandler) GetNextBatches(batchCount int) []*EntryScanResult {
	entryScanResults := []*EntryScanResult{}
	for ii := 0; ii < batchCount; ii++ {
		batchEvent := <-th.BatchedEntryChan
		nextEntryBatch := batchEvent.BatchedEntries
		entryScanResults = append(entryScanResults, &EntryScanResult{
			EntryBatch:     nextEntryBatch,
			IsMempool:      batchEvent.IsMempool,
			BatchesScanned: 1,
			FlushId:        nextEntryBatch[0].FlushId,
			EncoderType:    nextEntryBatch[0].EncoderType,
			OperationType:  nextEntryBatch[0].OperationType,
		})
	}
	return entryScanResults
}

// WaitForMatchingEntryBatch waits for an entry batch with the given encoder type, operation type, or flush id to appear in the BatchedEntryChan.
func (th *TestHandler) WaitForMatchingEntryBatch(
	targetEncoderType *lib.EncoderType,
	targetOpType *lib.StateSyncerOperationType,
	currentFlushId *uuid.UUID) (*EntryScanResult, error) {
	// Track the number of batches we've scanned.
	batchesScanned := 0

	// Continue retrieving entries from the BatchedEntryChan until we find the transaction hash.
	for batchEvent := range th.BatchedEntryChan {
		batchesScanned += 1

		nextEntryBatch := batchEvent.BatchedEntries

		encoderType := nextEntryBatch[0].EncoderType
		operationType := nextEntryBatch[0].OperationType
		flushId := nextEntryBatch[0].FlushId

		if (targetEncoderType == nil || encoderType == *targetEncoderType) &&
			(targetOpType == nil || operationType == *targetOpType) &&
			(currentFlushId == nil || *currentFlushId != flushId) {
			return &EntryScanResult{
				BatchesScanned: batchesScanned,
				EntryBatch:     nextEntryBatch,
				IsMempool:      batchEvent.IsMempool,
				FlushId:        nextEntryBatch[0].FlushId,
				EncoderType:    encoderType,
				OperationType:  operationType,
			}, nil
		}
	}
	return nil, fmt.Errorf("WaitForMatchingEntryBatch: Entry not found in entry batch")
}

// WaitForTxnHash waits for a transaction with the given hash to appear in the BatchedEntryChan.
func (th *TestHandler) WaitForTxnHash(txnHash string) (*EntryScanResult, error) {
	// Track the number of batches we've scanned.
	batchesScanned := 0

	// Continue retrieving entries from the BatchedEntryChan until we find the transaction hash.
	for batchEvent := range th.BatchedEntryChan {
		batchesScanned += 1

		nextEntryBatch := batchEvent.BatchedEntries

		//fmt.Printf("Skipping over encoder: %d|%d|%t|%s\n", nextEntryBatch[0].EncoderType, nextEntryBatch[0].OperationType, batchEvent.IsMempool, nextEntryBatch[0].FlushId.String())
		//if nextEntryBatch[0].EncoderType == lib.EncoderTypeBalanceEntry {
		//	fmt.Printf("Skipping over encoder: %d|%d|%t|%s\n", nextEntryBatch[0].EncoderType, nextEntryBatch[0].OperationType, batchEvent.IsMempool, nextEntryBatch[0].FlushId.String())
		//}

		txns, err := ParseTransactionsFromEntryBatch(nextEntryBatch, th.Params)
		if err != nil {
			return nil, errors.Wrapf(err, "WaitForTxnHash: Problem parsing transactions from entry batch")
		}

		for ii, txn := range txns {
			if txn.TransactionHash == txnHash {
				// Return the transaction, and all the following transactions.
				return &EntryScanResult{
					EntryBatch:     nextEntryBatch,
					BatchesScanned: batchesScanned,
					IsMempool:      batchEvent.IsMempool,
					FlushId:        nextEntryBatch[0].FlushId,
					EncoderType:    nextEntryBatch[0].EncoderType,
					OperationType:  nextEntryBatch[0].OperationType,
					Txn:            txn,
					RemainingTxns:  txns[ii+1:],
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("WaitForTxnHash: Transaction not found in entry batch")
}

func ParseTransactionsFromEntryBatch(entryBatch []*lib.StateChangeEntry, params *lib.DeSoParams) ([]*entries.PGTransactionEntry, error) {
	encoderType := entryBatch[0].EncoderType
	operationType := entryBatch[0].OperationType

	txns := []*entries.PGTransactionEntry{}

	if operationType == lib.DbOperationTypeDelete {
		return txns, nil
	}

	for _, entry := range entryBatch {
		if encoderType == lib.EncoderTypeBlock {
			blockTxns, err := BlockToTransactionEntries(entry.Encoder.(*lib.MsgDeSoBlock), entry.KeyBytes, params)
			if err != nil {
				return nil, errors.Wrapf(err, "ParseTransactionsFromEntryBatch: Problem converting block entry to transaction entries")
			}
			txns = append(txns, blockTxns...)
		} else if entry.Block != nil {
			blockTxns, err := BlockToTransactionEntries(entry.Block, entry.KeyBytes, params)
			if err != nil {
				return nil, errors.Wrapf(err, "ParseTransactionsFromEntryBatch: Problem converting block property to transaction entries")
			}
			txns = append(txns, blockTxns...)
		}
	}

	return txns, nil
}

func BlockToTransactionEntries(block *lib.MsgDeSoBlock, keyBytes []byte, params *lib.DeSoParams) ([]*entries.PGTransactionEntry, error) {
	blockEntry, _ := entries.BlockEncoderToPGStruct(block, keyBytes, params)
	txns := []*entries.PGTransactionEntry{}
	for ii, txn := range block.Txns {
		indexInBlock := uint64(ii)
		pgTxn, err := entries.TransactionEncoderToPGStruct(
			txn,
			&indexInBlock,
			blockEntry.BlockHash,
			blockEntry.Height,
			blockEntry.Timestamp,
			nil,
			nil,
			params,
		)
		if err != nil {
			return txns, errors.Wrapf(
				err,
				"entries.transformAndBulkInsertTransactionEntry: Problem converting transaction to PG struct",
			)
		}
		txns = append(txns, pgTxn)
	}
	return txns, nil
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
