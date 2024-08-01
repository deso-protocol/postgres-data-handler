package tests

import (
	"context"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/deso-protocol/backend/routes"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/postgres-data-handler/entries"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

// TODO: Make it so that the test only wipes out one of the limit orders.
// TODO: Test with a market order.

func TestDaoCoinLimitOrderFullOrderFillAtomic(t *testing.T) {
	t.Parallel()

	testCount := 10

	testUserCount := testCount * 3

	testConfig, err := SetupTestEnvironment(testUserCount, fmt.Sprintf("%s%d", t.Name(), rand.Intn(math.MaxInt16)))
	defer CleanupTestEnvironment(t, testConfig)
	require.NoError(t, err)

	nodeClient := testConfig.NodeClient

	firstTxValidatorFunc := func(t *testing.T, users []*TestUser, limitOrders []*entries.PGDaoCoinLimitOrderEntry) {
		require.Len(t, limitOrders, 1)
		limitOrder := limitOrders[0]
		require.Equal(t, users[0].PublicKeyBase58, limitOrder.SellingDaoCoinCreatorPkid)
		require.Equal(t, users[2].PublicKeyBase58, limitOrder.TransactorPkid)
		require.Equal(t, uint256.NewInt().SetUint64(1e9).String(), limitOrder.QuantityToFillInBaseUnitsHex)
	}

	secondTxValidatorFunc := func(t *testing.T, users []*TestUser, limitOrders []*entries.PGDaoCoinLimitOrderEntry) {
		require.Len(t, limitOrders, 0)
	}

	thirdTxValidatorFunc := func(t *testing.T, users []*TestUser, askOrders []*entries.PGDaoCoinLimitOrderEntry, bidOrders []*entries.PGDaoCoinLimitOrderEntry) {
		require.Len(t, bidOrders, 0)
		require.Len(t, askOrders, 0)
	}

	var wg sync.WaitGroup
	wg.Add(testCount)

	// Execute the DaoCoinLimitOrderTestExecution test in parallel, passing in the node client and the subset of users to use.
	for i := 0; i < testCount; i++ {
		go func(i int) {
			defer wg.Done()
			DaoCoinLimitOrderTestExecution(t, nodeClient, false,
				testConfig.TestUsers[i*3:(i+1)*3],
				1, 1, routes.DAOCoinLimitOrderFillTypeGoodTillCancelled, routes.DAOCoinLimitOrderOperationTypeStringASK,
				1, 1, routes.DAOCoinLimitOrderFillTypeGoodTillCancelled, routes.DAOCoinLimitOrderOperationTypeStringBID,
				firstTxValidatorFunc, secondTxValidatorFunc, thirdTxValidatorFunc)
		}(i)
	}

	// Wait for all goroutines to complete.
	wg.Wait()
}

func TestDaoCoinLimitOrderFullOrderFillSequentialBlocks(t *testing.T) {
	t.Parallel()

	testCount := 5

	testUserCount := testCount * 3

	testConfig, err := SetupTestEnvironment(testUserCount, fmt.Sprintf("%s%d", t.Name(), rand.Intn(math.MaxInt16)))
	defer CleanupTestEnvironment(t, testConfig)
	require.NoError(t, err)

	nodeClient := testConfig.NodeClient

	firstTxValidatorFunc := func(t *testing.T, users []*TestUser, limitOrders []*entries.PGDaoCoinLimitOrderEntry) {
		require.Len(t, limitOrders, 1)
		limitOrder := limitOrders[0]
		require.Equal(t, users[0].PublicKeyBase58, limitOrder.SellingDaoCoinCreatorPkid)
		require.Equal(t, users[2].PublicKeyBase58, limitOrder.TransactorPkid)
		require.Equal(t, uint256.NewInt().SetUint64(1e9).String(), limitOrder.QuantityToFillInBaseUnitsHex)
	}

	secondTxValidatorFunc := func(t *testing.T, users []*TestUser, limitOrders []*entries.PGDaoCoinLimitOrderEntry) {
		require.Len(t, limitOrders, 0)
	}

	thirdTxValidatorFunc := func(t *testing.T, users []*TestUser, askOrders []*entries.PGDaoCoinLimitOrderEntry, bidOrders []*entries.PGDaoCoinLimitOrderEntry) {
		require.Len(t, bidOrders, 0)
		require.Len(t, askOrders, 0)
	}

	var wg sync.WaitGroup
	wg.Add(testCount)

	// Execute the DaoCoinLimitOrderTestExecution test in parallel, passing in the node client and the subset of users to use.
	for i := 0; i < testCount; i++ {
		go func(i int) {
			defer wg.Done()
			DaoCoinLimitOrderTestExecution(t, nodeClient, true,
				testConfig.TestUsers[i*3:(i+1)*3],
				1, 1, routes.DAOCoinLimitOrderFillTypeGoodTillCancelled, routes.DAOCoinLimitOrderOperationTypeStringASK,
				1, 1, routes.DAOCoinLimitOrderFillTypeGoodTillCancelled, routes.DAOCoinLimitOrderOperationTypeStringBID,
				firstTxValidatorFunc, secondTxValidatorFunc, thirdTxValidatorFunc)
		}(i)
	}

	// Wait for all goroutines to complete.
	wg.Wait()
}

func SetupDaoCoinTest(t *testing.T, nodeClient *NodeClient, coinUser *TestUser, recipientUser *TestUser) {

	// Create the data structures necessary for executing the transactions atomically.
	var txns []*lib.MsgDeSoTxn
	var correspondingOwnerPublicKeysBase58Check []string
	var correspondingDerivedPublicKeysBase58Check []string
	var correspondingOptionalPrivKeys []*btcec.PrivateKey
	var correspondingIsDerived []bool

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

	daoCoinRes, _, err := nodeClient.DAOCoins(mintDaoCoinReq, coinUser.PrivateKey, false, false)
	require.NoError(t, err)

	// Append the unsubmitted transactions.
	txns = append(txns, daoCoinRes.Transaction)
	correspondingOwnerPublicKeysBase58Check = append(correspondingOwnerPublicKeysBase58Check, coinUser.PublicKeyBase58)
	correspondingDerivedPublicKeysBase58Check = append(correspondingDerivedPublicKeysBase58Check, "")
	correspondingOptionalPrivKeys = append(correspondingOptionalPrivKeys, coinUser.PrivateKey)
	correspondingIsDerived = append(correspondingIsDerived, false)

	// Transfer some coins to the ask user.
	daoCoinTransferRequest := &routes.TransferDAOCoinRequest{
		SenderPublicKeyBase58Check:             coinUser.PublicKeyBase58,
		ProfilePublicKeyBase58CheckOrUsername:  coinUser.PublicKeyBase58,
		ReceiverPublicKeyBase58CheckOrUsername: recipientUser.PublicKeyBase58,
		DAOCoinToTransferNanos:                 *uint256.NewInt().SetUint64(1e10),
		MinFeeRateNanosPerKB:                   FeeRateNanosPerKB,
		OptionalPrecedingTransactions:          txns,
	}

	transferDaoCoinRes, _, err := nodeClient.TransferDAOCoins(daoCoinTransferRequest, coinUser.PrivateKey, false, false)
	require.NoError(t, err)

	// Append the unsubmitted transactions.
	txns = append(txns, transferDaoCoinRes.Transaction)
	correspondingOwnerPublicKeysBase58Check = append(correspondingOwnerPublicKeysBase58Check, coinUser.PublicKeyBase58)
	correspondingDerivedPublicKeysBase58Check = append(correspondingDerivedPublicKeysBase58Check, "")
	correspondingOptionalPrivKeys = append(correspondingOptionalPrivKeys, coinUser.PrivateKey)
	correspondingIsDerived = append(correspondingIsDerived, false)

	// Submit the transactions atomically.
	submitRes, err := nodeClient.SignAndSubmitTxnsAtomically(txns, correspondingOwnerPublicKeysBase58Check, correspondingDerivedPublicKeysBase58Check, correspondingOptionalPrivKeys, correspondingIsDerived)
	require.NoError(t, err)

	// Wait for the mint transaction to be mined.
	_, err = nodeClient.WaitForTxnHash(submitRes.TxnHashHex, true)
	require.NoError(t, err)
}

func DaoCoinLimitOrderTestExecution(t *testing.T, nodeClient *NodeClient, signAndSubmitTxn bool, users []*TestUser,
	askPrice uint64, askQuantity uint64, askFillType routes.DAOCoinLimitOrderFillTypeString,
	askOperationType routes.DAOCoinLimitOrderOperationTypeString,
	bidPrice uint64, bidQuantity uint64, bidFillType routes.DAOCoinLimitOrderFillTypeString,
	bidOperationType routes.DAOCoinLimitOrderOperationTypeString,
	firstTransactionValidationFunc func(*testing.T, []*TestUser, []*entries.PGDaoCoinLimitOrderEntry),
	secondTransactionValidationFunc func(*testing.T, []*TestUser, []*entries.PGDaoCoinLimitOrderEntry),
	thirdTransactionValidationFunc func(*testing.T, []*TestUser, []*entries.PGDaoCoinLimitOrderEntry, []*entries.PGDaoCoinLimitOrderEntry),
) {
	coinUser := users[0]
	bidUser := users[1]
	askUser := users[2]

	SetupDaoCoinTest(t, nodeClient, coinUser, askUser)

	// Create the data structures necessary for executing the transactions atomically.
	var txns []*lib.MsgDeSoTxn
	var correspondingOwnerPublicKeysBase58Check []string
	var correspondingDerivedPublicKeysBase58Check []string
	var correspondingOptionalPrivKeys []*btcec.PrivateKey
	var correspondingIsDerived []bool

	// Create a limit ask order for the ask user.

	daoCoinAskReq := &routes.DAOCoinLimitOrderCreationRequest{
		TransactorPublicKeyBase58Check:            askUser.PublicKeyBase58,
		BuyingDAOCoinCreatorPublicKeyBase58Check:  routes.DESOCoinIdentifierString,
		SellingDAOCoinCreatorPublicKeyBase58Check: coinUser.PublicKeyBase58,
		Price:                strconv.FormatUint(askPrice, 10),
		Quantity:             ConvertToBaseUnitString(askQuantity),
		OperationType:        askOperationType,
		FillType:             askFillType,
		MinFeeRateNanosPerKB: FeeRateNanosPerKB,
	}

	limitOrderRes, txn, err := nodeClient.DaoCoinLimitOrderCreate(daoCoinAskReq, askUser.PrivateKey, false, signAndSubmitTxn)
	require.NoError(t, err)

	// Append the unsubmitted transactions.
	txns = append(txns, limitOrderRes.Transaction)
	correspondingOwnerPublicKeysBase58Check = append(correspondingOwnerPublicKeysBase58Check, askUser.PublicKeyBase58)
	correspondingDerivedPublicKeysBase58Check = append(correspondingDerivedPublicKeysBase58Check, "")
	correspondingOptionalPrivKeys = append(correspondingOptionalPrivKeys, askUser.PrivateKey)
	correspondingIsDerived = append(correspondingIsDerived, false)

	if signAndSubmitTxn {
		// Wait for the mint transaction to be mined.
		_, err = nodeClient.WaitForTxnHash(txn.TxnHashHex, true)
		require.NoError(t, err)

		fmt.Printf("Executing first limit order transaction: %+v\n", limitOrderRes)
		fmt.Printf("Executing first limit order transaction: %+v\n", daoCoinAskReq)

		// Retrieve the ask order, make sure it looks correct.
		limitOrders := []*entries.PGDaoCoinLimitOrderEntry{}
		err = nodeClient.StateSyncerDB.
			NewSelect().
			Model(&limitOrders).
			Where("selling_dao_coin_creator_pkid = ?", coinUser.PublicKeyBase58).
			Where("transactor_pkid = ?", askUser.PublicKeyBase58).
			Where("operation_type = ?", lib.DAOCoinLimitOrderOperationTypeASK).
			Scan(context.Background())
		require.NoError(t, err)
		if firstTransactionValidationFunc != nil {
			firstTransactionValidationFunc(t, users, limitOrders)
		}
	}

	// Create a buy order for the coin user.
	daoCoinBidReq := &routes.DAOCoinLimitOrderCreationRequest{
		TransactorPublicKeyBase58Check:            bidUser.PublicKeyBase58,
		BuyingDAOCoinCreatorPublicKeyBase58Check:  coinUser.PublicKeyBase58,
		SellingDAOCoinCreatorPublicKeyBase58Check: routes.DESOCoinIdentifierString,
		Price:                strconv.FormatUint(bidPrice, 10),
		Quantity:             ConvertToBaseUnitString(bidQuantity),
		OperationType:        bidOperationType,
		FillType:             bidFillType,
		MinFeeRateNanosPerKB: FeeRateNanosPerKB,
	}

	limitOrderRes2, txn, err := nodeClient.DaoCoinLimitOrderCreate(daoCoinBidReq, bidUser.PrivateKey, false, signAndSubmitTxn)
	require.NoError(t, err)

	// Append the unsubmitted transactions.
	txns = append(txns, limitOrderRes2.Transaction)
	correspondingOwnerPublicKeysBase58Check = append(correspondingOwnerPublicKeysBase58Check, bidUser.PublicKeyBase58)
	correspondingDerivedPublicKeysBase58Check = append(correspondingDerivedPublicKeysBase58Check, "")
	correspondingOptionalPrivKeys = append(correspondingOptionalPrivKeys, bidUser.PrivateKey)
	correspondingIsDerived = append(correspondingIsDerived, false)

	if signAndSubmitTxn {
		// Wait for the mint transaction to be mined.
		_, err = nodeClient.WaitForTxnHash(txn.TxnHashHex, true)
		require.NoError(t, err)

		// Confirm that the buy order was executed, and that the dao coin limit order table was updated.
		limitOrders := []*entries.PGDaoCoinLimitOrderEntry{}
		err = nodeClient.StateSyncerDB.
			NewSelect().
			Model(&limitOrders).
			Where("buying_dao_coin_creator_pkid = ?", coinUser.PublicKeyBase58).
			Where("transactor_pkid = ?", bidUser.PublicKeyBase58).
			Where("operation_type = ?", lib.DAOCoinLimitOrderOperationTypeBID).
			Scan(context.Background())
		require.NoError(t, err)
		if secondTransactionValidationFunc != nil {
			secondTransactionValidationFunc(t, users, limitOrders)
		}
	}

	// If submitting the orders atomically, submit the bundle and wait for it to be mined.
	if !signAndSubmitTxn {
		submitRes, err := nodeClient.SignAndSubmitTxnsAtomically(txns, correspondingOwnerPublicKeysBase58Check, correspondingDerivedPublicKeysBase58Check, correspondingOptionalPrivKeys, correspondingIsDerived)
		require.NoError(t, err)

		_, err = nodeClient.WaitForTxnHash(submitRes.TxnHashHex, true)
		require.NoError(t, err)
	}

	// Retrieve the bid and ask orders, make sure it looks correct.
	askLimitOrders := []*entries.PGDaoCoinLimitOrderEntry{}
	err = nodeClient.StateSyncerDB.
		NewSelect().
		Model(&askLimitOrders).
		Where("selling_dao_coin_creator_pkid = ?", coinUser.PublicKeyBase58).
		Where("transactor_pkid = ?", askUser.PublicKeyBase58).
		Where("operation_type = ?", lib.DAOCoinLimitOrderOperationTypeASK).
		Scan(context.Background())
	require.NoError(t, err)

	bidLimitOrders := []*entries.PGDaoCoinLimitOrderEntry{}
	err = nodeClient.StateSyncerDB.
		NewSelect().
		Model(&bidLimitOrders).
		Where("buying_dao_coin_creator_pkid = ?", coinUser.PublicKeyBase58).
		Where("transactor_pkid = ?", bidUser.PublicKeyBase58).
		Where("operation_type = ?", lib.DAOCoinLimitOrderOperationTypeBID).
		Scan(context.Background())
	require.NoError(t, err)

	if thirdTransactionValidationFunc != nil {
		thirdTransactionValidationFunc(t, users, askLimitOrders, bidLimitOrders)
	}
}
