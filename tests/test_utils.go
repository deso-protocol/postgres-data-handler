package tests

import (
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	be_routes "github.com/deso-protocol/backend/routes"
	"github.com/deso-protocol/backend/scripts/tools/toolslib"
	"github.com/deso-protocol/core/lib"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	POSBlockHeight = 30
)

type TestConfig struct {
	NodeClient *NodeClient
	TestUsers  []*TestUser
}

func SetupTestEnvironment(testUserCount int, password string, setupDb bool) (*TestConfig, error) {
	SetupFlags("../.env")
	stateSyncerPgUri, nodeUrl, logQueries := GetConfigValues()

	nodeClient, err := NewNodeClient(nodeUrl, stateSyncerPgUri, &lib.DeSoTestnetParams, logQueries, setupDb)
	if err != nil {
		glog.Fatalf("Error creating node client: %v", err)
	}

	// Wait for the node to be initialized to POS.
	nodeClient.WaitForBlockHeight(POSBlockHeight)

	// Create and fund test users.
	unsubmittedTestTxnBundles, testUsers, err := CreateAndFundTestUsers(nodeClient, testUserCount, password, false)
	if err != nil {
		glog.Fatalf("Error creating and funding test users: %v", err)
	}

	// Before we can sign and submit the transactions, we need to ensure that the max txn size is large enough to
	// accommodate the transactions we are about to submit. Just max that sucker out so we can have lots of headroom here.
	//_, err = IncreaseMaxTxnByteSize(nodeClient, testUsers[0], lib.MaxMaxTxnSizeBytes, true)
	//if err != nil {
	//	return nil, err
	//}

	var txns []*lib.MsgDeSoTxn
	var correspondingOwnerPublicKeysBase58Check []string
	var correspondingDerivedPublicKeysBase58Check []string
	var correspondingOptionalPrivKeys []*btcec.PrivateKey
	var correspondingIsDerived []bool

	for _, txnBundle := range unsubmittedTestTxnBundles {
		txns = append(txns, txnBundle.UnsubmittedTxn)
		correspondingOwnerPublicKeysBase58Check = append(correspondingOwnerPublicKeysBase58Check, txnBundle.CorrespondingOwnerPublicKeyBase58Check)
		correspondingDerivedPublicKeysBase58Check = append(correspondingDerivedPublicKeysBase58Check, txnBundle.CorrespondingDerivedPublicKeyBase58Check)
		correspondingOptionalPrivKeys = append(correspondingOptionalPrivKeys, txnBundle.CorrespondingPrivateKey)
		correspondingIsDerived = append(correspondingIsDerived, txnBundle.CorrespondingIsDerived)
	}

	// Sign and submit the transactions.
	submitRes, err := nodeClient.SignAndSubmitTxnsAtomically(txns, correspondingOwnerPublicKeysBase58Check, correspondingDerivedPublicKeysBase58Check, correspondingOptionalPrivKeys, correspondingIsDerived)
	if err != nil {
		return nil, err
	}

	if nodeClient.StateSyncerDB != nil {
		_, err = nodeClient.WaitForTxnHash(submitRes.TxnHashHex, true)
	}

	return &TestConfig{
		NodeClient: nodeClient,
		TestUsers:  testUsers,
	}, nil
}

func IncreaseMaxTxnByteSize(nodeClient *NodeClient, user *TestUser, byteSize uint64, signAndSubmitTxn bool) (*lib.MsgDeSoTxn, error) {
	updateGlobalParamsReq := &be_routes.UpdateGlobalParamsRequest{
		UpdaterPublicKeyBase58Check:    user.PublicKeyBase58,
		MaxBlockSizeBytesPoS:           byteSize,
		MaxTxnSizeBytesPoS:             byteSize,
		MinimumNetworkFeeNanosPerKB:    100,
		FeeBucketGrowthRateBasisPoints: 1000,
		MinFeeRateNanosPerKB:           FeeRateNanosPerKB,
	}

	updateGlobalParamsRes, _, err := nodeClient.UpdateGlobalParams(updateGlobalParamsReq, user.PrivateKey, false, signAndSubmitTxn)
	if err != nil {
		return nil, err
	}

	return updateGlobalParamsRes.Transaction, nil
}

func CleanupTestEnvironment(t *testing.T, testConfig *TestConfig) {
	// Close db connections.
	if testConfig != nil && testConfig.NodeClient != nil && testConfig.NodeClient.StateSyncerDB != nil {
		// Close state syncer db.
		if testConfig.NodeClient.StateSyncerDB != nil {
			err := testConfig.NodeClient.StateSyncerDB.Close()
			require.NoError(t, err)
		}
	}
	return
}

// Only configure flags once. This constraint is needed in the case of parallel tests.
var once sync.Once

func SetupFlags(envLocation string) {
	once.Do(func() {
		// Set glog flags
		flag.Set("log_dir", viper.GetString("log_dir"))
		flag.Set("v", viper.GetString("glog_v"))
		flag.Set("vmodule", viper.GetString("glog_vmodule"))
		flag.Set("alsologtostderr", "true")
		flag.Parse()
		glog.CopyStandardLogTo("INFO")
		viper.SetConfigFile(envLocation)
		viper.ReadInConfig()
		viper.AutomaticEnv()
	})
}

type TestUser struct {
	UserName                      string
	SeedBytes                     []byte
	SeedPhrase                    string
	AccountIdx                    uint32
	PublicKey                     *btcec.PublicKey
	PrivateKey                    *btcec.PrivateKey
	DerivedDefaultPublicKey       *btcec.PublicKey
	DerivedDefaultPublicKeyBase58 string
	DerivedDefaultPrivateKey      *btcec.PrivateKey
	PublicKeyBase58               string
	JWT                           string
}

// Create a new user for testing. The password and index passed here can be used in order to generate multiple different
// login credentials for a given seed.
func CreateTestUser(starterSeed string, password string, index uint32, desoParams *lib.DeSoParams, nodeClient *NodeClient) (*TestUser, *UnsubmittedTestTxnBundle, error) {
	seedBytes, err := bip39.NewSeedWithErrorChecking(starterSeed, password)
	if err != nil {
		return nil, nil, err
	}
	publicKey, privateKey, _, err := ComputeKeysFromSeedWithNet(seedBytes, index, desoParams)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBase58 := lib.Base58CheckEncode(publicKey.SerializeCompressed(), false, desoParams)
	jwtString, err := toolslib.GenerateJWTToken(privateKey)
	if err != nil {
		return nil, nil, err
	}

	derivedDefaultPrivKey, derivedDefaultPubKey := DeriveDefaultMessagingKey(privateKey, AccessGroupDefaultKeyName)
	derivedDefaultPubKeyBase58 := PubKeyToBase58(derivedDefaultPubKey, desoParams)

	userName := fmt.Sprintf("test%d", rand.Intn(math.MaxInt16))

	testUser := &TestUser{
		SeedBytes:                     seedBytes,
		SeedPhrase:                    starterSeed,
		AccountIdx:                    index,
		PublicKey:                     publicKey,
		PrivateKey:                    privateKey,
		PublicKeyBase58:               publicKeyBase58,
		DerivedDefaultPublicKeyBase58: derivedDefaultPubKeyBase58,
		DerivedDefaultPublicKey:       derivedDefaultPubKey,
		DerivedDefaultPrivateKey:      derivedDefaultPrivKey,
		JWT:                           jwtString,
	}

	txnBundle, err := CreateProfileForTestUser(nodeClient, testUser, userName)
	if err != nil {
		return nil, nil, err
	}

	return testUser, txnBundle, nil
}

func CreateProfileForTestUser(nodeClient *NodeClient, testUser *TestUser, username string) (*UnsubmittedTestTxnBundle, error) {
	// Construct request object.
	createProfileRequest := &be_routes.UpdateProfileRequest{
		UpdaterPublicKeyBase58Check: testUser.PublicKeyBase58,
		ProfilePublicKeyBase58Check: testUser.PublicKeyBase58,
		NewUsername:                 username,
		NewDescription:              fmt.Sprintf("This is a test profile for user %s", username),
		NewProfilePic:               DefaultProfilePicBase64,
		NewStakeMultipleBasisPoints: 1.25 * 100 * 100,
		IsHidden:                    false,
		ExtraData:                   nil,
		MinFeeRateNanosPerKB:        FeeRateNanosPerKB,
	}
	updateProfResponse, _, err := nodeClient.UpdateProfile(createProfileRequest, testUser.PrivateKey, false, false)
	if err != nil {
		return nil, err
	}

	unsubmittedTestTxnBundle := &UnsubmittedTestTxnBundle{
		UnsubmittedTxn:                           updateProfResponse.Transaction,
		CorrespondingOwnerPublicKeyBase58Check:   testUser.PublicKeyBase58,
		CorrespondingDerivedPublicKeyBase58Check: "",
		CorrespondingPrivateKey:                  testUser.PrivateKey,
		CorrespondingIsDerived:                   false,
	}

	return unsubmittedTestTxnBundle, err
}

// Used for streamlining atomics in testing.
type UnsubmittedTestTxnBundle struct {
	UnsubmittedTxn                           *lib.MsgDeSoTxn
	CorrespondingOwnerPublicKeyBase58Check   string
	CorrespondingDerivedPublicKeyBase58Check string
	CorrespondingPrivateKey                  *btcec.PrivateKey
	CorrespondingIsDerived                   bool
}

func CreateAndFundTestUsers(
	nodeClient *NodeClient,
	testUserCount int,
	password string,
	signAndSubmitTxn bool,
) (
	unsubmittedTestTxnBundles []*UnsubmittedTestTxnBundle,
	returnTestUsers []*TestUser,
	err error,
) {
	// Derive starter account credentials and attach to the test config.
	// This is the account that will have starter DESO.
	starterAccountSeed := viper.GetString("TEST_STARTER_DESO_SEED")
	starterUser, txnBundle, err := CreateTestUser(starterAccountSeed, "", 0, nodeClient.DeSoParams, nodeClient)
	if err != nil {
		return nil, nil, err
	}

	unsubmittedTestTxnBundles = append(unsubmittedTestTxnBundles, txnBundle)

	// Preallocate the slice
	testUsers := make([]*TestUser, testUserCount+2)
	testUsers[0] = starterUser

	// Utilize a wait group to create & fund the users in parallel.
	var wg sync.WaitGroup

	type SafeUnsubmittedTestTxnBundles struct {
		UnsubmittedTestTxnBundles []*UnsubmittedTestTxnBundle
		sync.Mutex
	}
	var safeUnsubmittedTestTxnBundles SafeUnsubmittedTestTxnBundles

	for ii := 0; ii < testUserCount+1; ii++ {
		wg.Add(1)
		go func(ii int) {
			defer wg.Done()

			testUser, unsubmittedTestUserTxnBundle, createErr := CreateTestUser(starterAccountSeed, password, uint32(ii), nodeClient.DeSoParams, nodeClient)
			if createErr != nil {
				err = createErr
				return
			}

			testUsers[ii+1] = testUser

			unsubmittedTestTxnBundle, fundErr := FundTestUser(
				nodeClient, starterUser, testUser.PublicKeyBase58, TestUserStarterNanos)
			if fundErr != nil {
				err = fundErr
				return
			}

			safeUnsubmittedTestTxnBundles.Lock()
			defer safeUnsubmittedTestTxnBundles.Unlock()
			safeUnsubmittedTestTxnBundles.UnsubmittedTestTxnBundles = append(
				safeUnsubmittedTestTxnBundles.UnsubmittedTestTxnBundles, unsubmittedTestTxnBundle)
			safeUnsubmittedTestTxnBundles.UnsubmittedTestTxnBundles = append(
				safeUnsubmittedTestTxnBundles.UnsubmittedTestTxnBundles, unsubmittedTestUserTxnBundle)
		}(ii)
	}
	// Wait for all users to be created.
	wg.Wait()

	return safeUnsubmittedTestTxnBundles.UnsubmittedTestTxnBundles, testUsers, err
}

// Use funds from the starter user to fund the test user.
func FundTestUser(nodeClient *NodeClient, senderUser *TestUser, recipientUserPublicKey string, amountNanos int64) (*UnsubmittedTestTxnBundle, error) {
	// Construct request object.
	sendDesoRequest := &be_routes.SendDeSoRequest{
		SenderPublicKeyBase58Check:   senderUser.PublicKeyBase58,
		RecipientPublicKeyOrUsername: recipientUserPublicKey,
		AmountNanos:                  amountNanos,
		MinFeeRateNanosPerKB:         FeeRateNanosPerKB,
	}

	// Create request to backend.
	sendDesoRes, _, err := nodeClient.SendDeso(sendDesoRequest, senderUser.PrivateKey, false, false)
	if err != nil {
		return nil, err
	}

	unsubmittedTestTxnBundle := &UnsubmittedTestTxnBundle{
		UnsubmittedTxn:                           sendDesoRes.Transaction,
		CorrespondingOwnerPublicKeyBase58Check:   senderUser.PublicKeyBase58,
		CorrespondingDerivedPublicKeyBase58Check: "",
		CorrespondingPrivateKey:                  senderUser.PrivateKey,
		CorrespondingIsDerived:                   false,
	}

	return unsubmittedTestTxnBundle, nil
}

func GetConfigValues() (string, string, bool) {
	dbHost := viper.GetString("DB_HOST")
	dbPort := viper.GetString("DB_PORT")
	dbUsername := viper.GetString("DB_USERNAME")
	dbPassword := viper.GetString("DB_PASSWORD")
	dbName := "postgres"

	pgURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&timeout=18000s", dbUsername, dbPassword, dbHost, dbPort, dbName)

	nodeUrl := viper.GetString("NODE_URL")

	logQueries := viper.GetBool("LOG_QUERIES")

	return pgURI, nodeUrl, logQueries
}

func ConvertToBaseUnitString(value uint64) string {
	// Perform the division and convert to float64
	decimalValue := float64(value) / 1e9
	// Convert the result to a string
	return strconv.FormatFloat(decimalValue, 'f', 9, 64)
}

func (nodeClient *NodeClient) WaitForBlockHeight(blockHeight uint32) {
	for {
		appState, _, err := nodeClient.GetAppState(&be_routes.GetAppStateRequest{PublicKeyBase58Check: ""}, nil, false)
		if err == nil && appState.BlockHeight >= blockHeight {
			return
		}
		// Sleep for a second before checking again.
		time.Sleep(1 * time.Second)
	}
}
