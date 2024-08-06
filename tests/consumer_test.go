package tests

import (
	"github.com/deso-protocol/backend/config"
	"github.com/deso-protocol/backend/routes"
	coreCmd "github.com/deso-protocol/core/cmd"
	"github.com/deso-protocol/core/lib"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
	"os"
	"testing"
)

const (
	globalStateSharedSecret = "abcdef"
)

func TestConsumer(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "ss")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	// Clean up the temporary directory after the test
	defer os.RemoveAll(tempDir)

	desoParams := &lib.DeSoTestnetParams
	starterSeed := "verb find card ship another until version devote guilt strong lemon six"
	seedBytes, err := bip39.NewSeedWithErrorChecking(starterSeed, "")
	require.NoError(t, err)
	publicKey, _, _, err := ComputeKeysFromSeedWithNet(seedBytes, 0, desoParams)
	require.NoError(t, err)

	publicKeyBase58 := lib.Base58CheckEncode(publicKey.SerializeCompressed(), false, desoParams)
	require.NoError(t, err)

	//_, nodeServer := newTestApiServer(t, publicKeyBase58, 17001)
	newTestApiServer(t, publicKeyBase58, 17001)
	require.NoError(t, err)

	//// Wait 2 seconds.
	//time.Sleep(2 * time.Second)
	//
	//// Pause blocksync.
	//nodeServer.StateChangeSyncer.PauseBlockSync = false
	//
	//time.Sleep(2 * time.Second)

	//nodeServer.PauseMempoolSync = true
}

func newTestApiServer(t *testing.T, starterUserPublicKeyBase58 string, apiPort uint16) (*routes.APIServer, *lib.Server) {
	// Create a badger db instance.
	badgerDB, badgerDir := routes.GetTestBadgerDb(t)

	// Set core node's config.
	coreConfig := coreCmd.LoadConfig()
	coreConfig.Params = &lib.DeSoTestnetParams
	coreConfig.DataDirectory = badgerDir
	coreConfig.Regtest = true
	coreConfig.TXIndex = false
	coreConfig.MinerPublicKeys = []string{starterUserPublicKeyBase58}
	coreConfig.NumMiningThreads = 1
	coreConfig.HyperSync = false
	coreConfig.MinFeerate = 2000
	coreConfig.LogDirectory = badgerDir
	coreConfig.StateChangeDir = "./ss/state-changes"

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
	apiServer.Start()

	t.Cleanup(func() {
		apiServer.Stop()
		node.Stop()
	})
	return apiServer, node.Server
}
