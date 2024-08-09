package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/deso-protocol/backend/routes"
	"github.com/deso-protocol/core/lib"
	"github.com/deso-protocol/postgres-data-handler/entries"
	initial_migrations "github.com/deso-protocol/postgres-data-handler/migrations/initial_migrations"
	post_sync_migrations "github.com/deso-protocol/postgres-data-handler/migrations/post_sync_migrations"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	TestUserStarterNanos      = 1e9
	FeeRateNanosPerKB         = 3000
	AccessGroupDefaultKeyName = "default-key"
	DefaultProfilePicBase64   = "data:image/png;base64,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"
)

type NodeClient struct {
	NodeURL       string
	StateSyncerDB *bun.DB
	DeSoParams    *lib.DeSoParams
}

func ConstructPgURI(
	dbHost string,
	dbPort string,
	dbName string,
	dbUser string,
	dbPassword string,
) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
}

func NewNodeClient(nodeURL string, pgURI string, desoParams *lib.DeSoParams, logQueries bool, setupDb bool) (*NodeClient, error) {
	var db *bun.DB

	if setupDb {

		// Open a PostgreSQL database.
		pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(pgURI)))
		if pgdb == nil {
			glog.Fatalf("Error connecting to postgres db at URI: %v", pgURI)
		}

		// Create a Bun db on top of postgres for querying.
		db = bun.NewDB(pgdb, pgdialect.New())
		db.SetConnMaxLifetime(0)
		db.SetMaxIdleConns(10)

		// Print all queries to stdout for debugging.
		if logQueries {
			db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
		}

		ctx := context.Background()

		// Apply db migrations.
		migrator := migrate.NewMigrator(db, initial_migrations.Migrations)
		if err := migrator.Init(ctx); err != nil {
			return nil, err
		}
		group, err := migrator.Migrate(ctx)
		if err != nil {
			return nil, err
		}

		migrator = migrate.NewMigrator(db, post_sync_migrations.Migrations)
		if err = migrator.Init(ctx); err != nil {
			return nil, err
		}
		group, err = migrator.Migrate(ctx)
		if err != nil {
			return nil, err
		}

		if logQueries {
			glog.Infof("Migrated to %s\n", group)
		}
	}

	nodeClient := &NodeClient{
		NodeURL:       nodeURL,
		StateSyncerDB: db,
		DeSoParams:    desoParams,
	}

	return nodeClient, nil
}

func (nodeClient *NodeClient) SignTransaction(
	privKey *btcec.PrivateKey,
	transactionHex string,
	isDerived bool,
) (
	_signedTxn *lib.MsgDeSoTxn,
	_err error,
) {

	// Get the transaction bytes from the request data.
	txnBytes, err := hex.DecodeString(transactionHex)
	if err != nil {
		return nil, fmt.Errorf("SignTransaction: Problem decoding transaction hex %v", err)
	}

	txn := &lib.MsgDeSoTxn{}
	if err = txn.FromBytes(txnBytes); err != nil {
		return nil, fmt.Errorf("SignTransaction: Problem deserializing transaction: %v", err)
	}

	// Sign the transaction with a derived key. Since the txn extraData must be modified,
	// we also get new transaction bytes, along with the signature.
	newTransactionBytes, txnSignatureBytes, err := lib.SignTransactionBytes(txnBytes, privKey, isDerived)

	parsedSignature, err := btcec.ParseDERSignature(txnSignatureBytes, btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("SignTransaction: Problem parsing signature: %v", err)
	}

	newTxn := &lib.MsgDeSoTxn{}
	if err = newTxn.FromBytes(newTransactionBytes); err != nil {
		return nil, fmt.Errorf("SignTransaction: Problem deserializing new transaction: %v", err)
	}

	newTxn.Signature.SetSignature(parsedSignature)
	return newTxn, nil
}

func (nodeClient *NodeClient) SubmitTransaction(txn *lib.MsgDeSoTxn) (*routes.SubmitTransactionResponse, error) {
	// The response will contain the new transaction bytes and a signature.
	signedTransactionHex, err := txn.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("SubmitTransaction: Problem serializing new transaction: %v", err)
	}
	signedTransactionHexStr := hex.EncodeToString(signedTransactionHex)

	jsonStr, err := json.Marshal(&routes.SubmitTransactionRequest{
		TransactionHex: signedTransactionHexStr,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "SubmitTransaction: Error serializing SubmitTransactionRequest: %v", err)
	}

	// Submit the transaction
	submitTransactionURL := fmt.Sprintf("%s/api/v0/submit-transaction",
		nodeClient.NodeURL)

	req, err := http.NewRequest("POST", submitTransactionURL, bytes.NewBuffer([]byte(jsonStr)))
	req.Header.Set("Origin", nodeClient.NodeURL)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	submitResp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer submitResp.Body.Close()

	body, _ := ioutil.ReadAll(submitResp.Body)
	if submitResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SubmitTransaction: Error Status returned "+
			"from %v: %v, %v", submitTransactionURL, submitResp.Status, string(body))
	}

	submitTransactionResponse := &routes.SubmitTransactionResponse{}
	submitTransactionDecoder := json.NewDecoder(bytes.NewReader(body))
	if err := submitTransactionDecoder.Decode(submitTransactionResponse); err != nil {
		return nil, fmt.Errorf(
			"SubmitTransaction: Error parsing JSON response: %v", err)
	}

	return submitTransactionResponse, nil
}

func (nodeClient *NodeClient) SignAndSubmitTransaction(
	privKey *btcec.PrivateKey,
	transactionHex string,
	isDerived bool,
) (
	_submitTransactionResponse *routes.SubmitTransactionResponse,
	_err error,
) {

	newTxn, err := nodeClient.SignTransaction(
		privKey,
		transactionHex,
		isDerived,
	)
	if err != nil {
		return nil, fmt.Errorf("SignAndSubmitTransaction: Error signing transaction: %v", err)
	}

	submitTransactionResponse, err := nodeClient.SubmitTransaction(newTxn)
	if err != nil {
		return nil, fmt.Errorf("SignAndSubmitTransaction: Error submitting transaction: %v", err)
	}

	return submitTransactionResponse, nil
}

func (nodeClient *NodeClient) SignAndSubmitTxnsAtomically(
	txns []*lib.MsgDeSoTxn,
	correspondingOwnerPublicKeysBase58Check []string,
	correspondingDerivedPublicKeysBase58Check []string,
	correspondingOptionalPrivKeys []*btcec.PrivateKey,
	correspondingIsDerived []bool,
) (
	_submitAtomicTransactionResponse *routes.SubmitAtomicTransactionResponse,
	_err error,
) {
	// Sanity check the input.
	if len(txns) != len(correspondingOwnerPublicKeysBase58Check) ||
		len(txns) != len(correspondingDerivedPublicKeysBase58Check) ||
		len(txns) != len(correspondingOptionalPrivKeys) ||
		len(txns) != len(correspondingIsDerived) {
		return nil,
			fmt.Errorf("SignAndSubmitTxnsAtomically: Number of transactions must match number of keys")
	}

	// First we must sign each transaction provided.
	// This seems incorrect, but the reason for this step is to ensure the
	// derived public key is set in the transaction extra data before constructing
	// the atomic transactions' wrapper.
	var txnsWithDerivedKeyExtraData []*lib.MsgDeSoTxn
	for idx, txn := range txns {
		// Encode the transaction to hex.
		txnBytes, err := txn.ToBytes(true)
		if err != nil {
			return nil,
				fmt.Errorf("SignAndSubmitTxnsAtomically: Problem serializing transaction: %v", err)
		}
		txnHex := hex.EncodeToString(txnBytes)

		// Sign the internal transaction.
		signedTxn, err := nodeClient.SignTransaction(
			correspondingOptionalPrivKeys[idx],
			txnHex,
			correspondingIsDerived[idx],
		)
		if err != nil {
			return nil,
				fmt.Errorf("SignAtomicTxns: Error signing transaction %d: %v", idx, err)
		}

		// Delete the signature field to ensure this doesn't leak to the network.
		signedTxn.Signature = lib.DeSoSignature{}
		txnsWithDerivedKeyExtraData = append(txnsWithDerivedKeyExtraData, signedTxn)
	}

	// Get the atomic txns wrapper from the provided transactions.
	createAtomicTxnsWrapperResponse, err := nodeClient.GetAtomicTxnsWrapper(txnsWithDerivedKeyExtraData, nil)
	if err != nil {
		return nil,
			fmt.Errorf("SignAndSubmitTxnsAtomically: Error getting atomic txns wrapper: %v", err)
	}

	// Sign and submit the atomic transactions.
	submitAtomicTransactionResponse, err := nodeClient.SignAndSubmitAtomicTxns(
		createAtomicTxnsWrapperResponse.Transaction,
		correspondingOptionalPrivKeys,
		correspondingIsDerived,
	)
	if err != nil {
		return nil,
			fmt.Errorf("SignAndSubmitTxnsAtomically: Error signing and submitting atomic transactions: %v", err)
	}
	return submitAtomicTransactionResponse, nil
}

func (nodeClient *NodeClient) SignAndSubmitAtomicTxns(
	atomicTxnsWrapper *lib.MsgDeSoTxn,
	correspondingOptionalPrivKeys []*btcec.PrivateKey,
	correspondingIsDerived []bool,
) (
	_submitAtomicTransactionResponse *routes.SubmitAtomicTransactionResponse,
	_err error,
) {
	// Sign the atomic transactions.
	if err := nodeClient.SignAtomicTxns(
		atomicTxnsWrapper,
		correspondingOptionalPrivKeys,
		correspondingIsDerived,
	); err != nil {
		return nil,
			fmt.Errorf("SignAndSubmitAtomicTxns: Error signing atomic transactions: %v", err)
	}

	// Submit the atomic transactions.
	submitTransactionResponse, err := nodeClient.SubmitAtomicTransaction(atomicTxnsWrapper, nil)
	if err != nil {
		return nil,
			fmt.Errorf("SignAndSubmitAtomicTxns: Error submitting atomic transactions: %v", err)
	}

	return submitTransactionResponse, nil
}

func (nodeClient *NodeClient) SignAtomicTxns(
	atomicTxnsWrapper *lib.MsgDeSoTxn,
	correspondingOptionalPrivKeys []*btcec.PrivateKey,
	correspondingIsDerived []bool,
) (
	_err error,
) {
	// Validate the atomicTxnsWrapper is the correct type.
	if atomicTxnsWrapper.TxnMeta.GetTxnType() != lib.TxnTypeAtomicTxnsWrapper {
		return fmt.Errorf("SignAtomicTxns: Transaction must be of type AtomicTxnsWrapper")
	}

	// Validate that we have the correct number of private keys to sign the entire atomicTxnsWrapper.
	if len(atomicTxnsWrapper.TxnMeta.(*lib.AtomicTxnsWrapperMetadata).Txns) != len(correspondingOptionalPrivKeys) {
		return fmt.Errorf("SignAtomicTxns: Number of private keys must " +
			"match number of transactions in atomicTxnsWrapper")
	}

	// Sign the corresponding transactions.
	for idx, txn := range atomicTxnsWrapper.TxnMeta.(*lib.AtomicTxnsWrapperMetadata).Txns {
		// Encode the transaction to hex.
		txnBytes, err := txn.ToBytes(true)
		if err != nil {
			return fmt.Errorf("SignAtomicTxns: Problem serializing transaction: %v", err)
		}
		txnHex := hex.EncodeToString(txnBytes)

		// Sign the internal transaction.
		signedTxn, err := nodeClient.SignTransaction(
			correspondingOptionalPrivKeys[idx],
			txnHex,
			correspondingIsDerived[idx],
		)
		if err != nil {
			return fmt.Errorf("SignAtomicTxns: Error signing transaction %d: %v", idx, err)
		}

		// Replace the transaction in the atomicTxnsWrapper with the signed transaction.
		atomicTxnsWrapper.TxnMeta.(*lib.AtomicTxnsWrapperMetadata).Txns[idx] = signedTxn
	}
	return nil
}

func (nodeClient *NodeClient) SubmitAtomicTransaction(
	atomicTxnsWrapper *lib.MsgDeSoTxn,
	optionalSignedInnerTransactionsHex []string,
) (
	*routes.SubmitAtomicTransactionResponse,
	error,
) {
	// Ensure the atomicTxnsWrapper is the right transaction type.
	if atomicTxnsWrapper.TxnMeta.GetTxnType() != lib.TxnTypeAtomicTxnsWrapper {
		return nil, fmt.Errorf("SubmitAtomicTransaction: Transaction must be of type AtomicTxnsWrapper")
	}

	// Encode the atomicTxnsWrapper to hex.
	atomicTxnsWrapperBytes, err := atomicTxnsWrapper.ToBytes(false)
	if err != nil {
		return nil, fmt.Errorf("SubmitAtomicTransaction: Problem serializing atomicTxnsWrapper: %v", err)
	}
	atomicTxnsWrapperHex := hex.EncodeToString(atomicTxnsWrapperBytes)

	// JSON marshal the request.
	jsonStr, err := json.Marshal(&routes.SubmitAtomicTransactionRequest{
		IncompleteAtomicTransactionHex: atomicTxnsWrapperHex,
		SignedInnerTransactionsHex:     optionalSignedInnerTransactionsHex,
	})
	if err != nil {
		return nil,
			errors.Wrapf(err, "SubmitAtomicTransaction: Error serializing SubmitAtomicTransactionRequest")
	}

	// Submit the transaction.
	submitAtomicTransactionURL := fmt.Sprintf("%s"+routes.RoutePathSubmitAtomicTransaction,
		nodeClient.NodeURL)
	req, err := http.NewRequest("POST", submitAtomicTransactionURL, bytes.NewBuffer([]byte(jsonStr)))
	req.Header.Set("Origin", nodeClient.NodeURL)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	submitResp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer submitResp.Body.Close()

	body, _ := ioutil.ReadAll(submitResp.Body)
	if submitResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SubmitTransaction: Error Status returned "+
			"from %v: %v, %v", submitAtomicTransactionURL, submitResp.Status, string(body))
	}

	// Decode the response.
	submitAtomicTransactionResponse := &routes.SubmitAtomicTransactionResponse{}
	submitAtomicTransactionDecoder := json.NewDecoder(bytes.NewReader(body))
	if err := submitAtomicTransactionDecoder.Decode(submitAtomicTransactionResponse); err != nil {
		return nil, fmt.Errorf("SubmitAtomicTransaction: Error parsing JSON response: %v", err)
	}
	return submitAtomicTransactionResponse, nil
}

func NodePostRequest[Req any, Resp any](
	nodeClient *NodeClient,
	routePath string,
	request Req,
	signAndSubmitTxn bool,
	privateKey *btcec.PrivateKey,
	transactionHexFieldName string,
	isDerived bool) (*Resp, *routes.SubmitTransactionResponse, error) {
	body, err := MakePostRequest(nodeClient.NodeURL, routePath, &request)
	if err != nil {
		return nil, nil, fmt.Errorf("Error hitting endpoint: %v", err)
	}

	var response Resp
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err = decoder.Decode(&response); err != nil {
		return nil, nil, fmt.Errorf("Error parsing JSON response: %v", err)
	}

	var txResp *routes.SubmitTransactionResponse

	if signAndSubmitTxn {
		// Reflection to check for TransactionHex field
		respValue := reflect.ValueOf(response)
		txHexField := respValue.FieldByName(transactionHexFieldName)
		if !txHexField.IsValid() {
			return nil, nil, fmt.Errorf("Response type does not have a TransactionHex field")
		}
		transactionHex := txHexField.String()

		submitTries := 5

		for i := 0; i < submitTries; i++ {
			if txResp, err = nodeClient.SignAndSubmitTransaction(
				privateKey, transactionHex, isDerived); err != nil {
				// Check to see if the error message includes "NEED_BLOCKS" anywhere in the error string. If so, wait, and retry.
				if strings.Contains(err.Error(), "NEED_BLOCKS") {
					time.Sleep(250 * time.Millisecond)
					continue
				}

				return nil, nil, fmt.Errorf("Error signing and submitting transaction: %v", err)
			} else {
				break
			}
		}

		if err != nil {
			return nil, nil, fmt.Errorf("Error signing and submitting transaction: %v", err)
		}
	}

	return &response, txResp, nil
}

func (nodeClient *NodeClient) GetAtomicTxnsWrapper(
	txns []*lib.MsgDeSoTxn,
	extraData map[string]string,
) (
	*routes.CreateAtomicTxnsWrapperResponse,
	error,
) {
	body, err := MakePostRequest(
		nodeClient.NodeURL,
		routes.RoutePathCreateAtomicTxnsWrapper,
		&routes.CreateAtomicTxnsWrapperRequest{
			Transactions:         txns,
			ExtraData:            extraData,
			MinFeeRateNanosPerKB: FeeRateNanosPerKB,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Error hitting endpoint: %v", err)
	}

	var response routes.CreateAtomicTxnsWrapperResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err = decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("Error parsing JSON response: %v", err)
	}

	return &response, nil
}

func (nodeClient *NodeClient) GetTxnInfo(request *routes.APITransactionInfoRequest) (*routes.APITransactionInfoResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.APITransactionInfoRequest, routes.APITransactionInfoResponse](nodeClient, routes.RoutePathAPITransactionInfo, *request, false, nil, "", false)
}

func (nodeClient *NodeClient) AuthorizeDerivedKey(
	request *routes.AuthorizeDerivedKeyRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.AuthorizeDerivedKeyResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.AuthorizeDerivedKeyRequest, routes.AuthorizeDerivedKeyResponse](nodeClient, routes.RoutePathAuthorizeDerivedKey, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) SubmitPost(request *routes.SubmitPostRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.SubmitPostResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.SubmitPostRequest, routes.SubmitPostResponse](nodeClient, routes.RoutePathSubmitPost, *request, true, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) GetSinglePost(request *routes.GetSinglePostRequest) (*routes.GetSinglePostResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.GetSinglePostRequest, routes.GetSinglePostResponse](nodeClient, routes.RoutePathGetSinglePost, *request, false, nil, "", false)
}

func (nodeClient *NodeClient) GetPostsForUser(request *routes.GetPostsForPublicKeyRequest) (*routes.GetPostsForPublicKeyResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.GetPostsForPublicKeyRequest, routes.GetPostsForPublicKeyResponse](nodeClient, routes.RoutePathGetPostsForPublicKey, *request, false, nil, "", false)
}

func (nodeClient *NodeClient) SendDeso(
	request *routes.SendDeSoRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.SendDeSoResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.SendDeSoRequest, routes.SendDeSoResponse](nodeClient, routes.RoutePathSendDeSo, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) TransferDaoCoin(
	request *routes.TransferDAOCoinRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.TransferDAOCoinResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.TransferDAOCoinRequest, routes.TransferDAOCoinResponse](nodeClient, routes.RoutePathTransferDAOCoin, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) DaoCoin(request *routes.DAOCoinRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.DAOCoinResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.DAOCoinRequest, routes.DAOCoinResponse](nodeClient, routes.RoutePathDAOCoin, *request, true, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) DaoCoinLimitOrderCreate(request *routes.DAOCoinLimitOrderCreationRequest, privateKey *btcec.PrivateKey, isDerived bool, signAndSubmitTxn bool) (*routes.DAOCoinLimitOrderResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.DAOCoinLimitOrderCreationRequest, routes.DAOCoinLimitOrderResponse](nodeClient, routes.RoutePathCreateDAOCoinLimitOrder, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) DaoCoinMarketOrderCreate(request *routes.DAOCoinMarketOrderCreationRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.DAOCoinLimitOrderResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.DAOCoinMarketOrderCreationRequest, routes.DAOCoinLimitOrderResponse](nodeClient, routes.RoutePathCreateDAOCoinMarketOrder, *request, true, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) DaoCoinLimitOrderCancel(request *routes.DAOCoinLimitOrderWithCancelOrderIDRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.DAOCoinLimitOrderResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.DAOCoinLimitOrderWithCancelOrderIDRequest, routes.DAOCoinLimitOrderResponse](nodeClient, routes.RoutePathCancelDAOCoinLimitOrder, *request, true, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) UpdateProfile(
	request *routes.UpdateProfileRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (*routes.UpdateProfileResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.UpdateProfileRequest, routes.UpdateProfileResponse](nodeClient, routes.RoutePathUpdateProfile, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) CreateAccessGroup(
	request *routes.CreateAccessGroupRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.CreateAccessGroupResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.CreateAccessGroupRequest, routes.CreateAccessGroupResponse](nodeClient, routes.RoutePathCreateAccessGroup, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) AddAccessGroupMember(
	request *routes.AddAccessGroupMembersRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.AddAccessGroupMembersResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.AddAccessGroupMembersRequest, routes.AddAccessGroupMembersResponse](nodeClient, routes.RoutePathAddAccessGroupMembers, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) RemoveAccessGroupMember(
	request *routes.AddAccessGroupMembersRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.AddAccessGroupMembersResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.AddAccessGroupMembersRequest, routes.AddAccessGroupMembersResponse](nodeClient, routes.RoutePathRemoveAccessGroupMembers, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) CreateUserAssociation(
	request *routes.CreateUserAssociationRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.AssociationTxnResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.CreateUserAssociationRequest, routes.AssociationTxnResponse](nodeClient, routes.RoutePathUserAssociations+"/create", *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) CreatePostAssociation(
	request *routes.CreatePostAssociationRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.AssociationTxnResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.CreatePostAssociationRequest, routes.AssociationTxnResponse](nodeClient, routes.RoutePathPostAssociations+"/create", *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) DeleteUserAssociation(
	request *routes.DeleteAssociationRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (
	*routes.AssociationTxnResponse,
	*routes.SubmitTransactionResponse,
	error,
) {
	return NodePostRequest[routes.DeleteAssociationRequest, routes.AssociationTxnResponse](nodeClient, routes.RoutePathUserAssociations+"/delete", *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) GetAccessGroupInfo(request *routes.GetAccessGroupInfoRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.AccessGroupEntryResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.GetAccessGroupInfoRequest, routes.AccessGroupEntryResponse](nodeClient, routes.RoutePathGetAccessGroupInfo, *request, false, privateKey, "", isDerived)
}

func (nodeClient *NodeClient) GetUsersStateless(request *routes.GetUsersStatelessRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.GetUsersResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.GetUsersStatelessRequest, routes.GetUsersResponse](nodeClient, routes.RoutePathGetUsersStateless, *request, false, privateKey, "", isDerived)
}

func (nodeClient *NodeClient) GetUsernamesForPublicKeys(publicKeyBase58Checks []string) (map[string]string, error) {
	response, _, err := nodeClient.GetUsersStateless(&routes.GetUsersStatelessRequest{
		PublicKeysBase58Check: publicKeyBase58Checks,
		IncludeBalance:        false,
		GetUnminedBalance:     false,
	}, nil, false)
	if err != nil || response == nil {
		return nil, fmt.Errorf("error getting users stateless: %v", err)
	}
	pkToUsernameMap := make(map[string]string)
	for _, user := range response.UserList {
		if user.ProfileEntryResponse == nil {
			continue
		}
		pkToUsernameMap[user.PublicKeyBase58Check] = user.ProfileEntryResponse.Username
	}
	return pkToUsernameMap, nil
}

func (nodeClient *NodeClient) GetAppState(request *routes.GetAppStateRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.GetAppStateResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.GetAppStateRequest, routes.GetAppStateResponse](nodeClient, routes.RoutePathGetAppState, *request, false, privateKey, "", isDerived)
}

func (nodeClient *NodeClient) IsHodlingPublicKey(request *routes.IsHodlingPublicKeyRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.IsHodlingPublicKeyResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.IsHodlingPublicKeyRequest, routes.IsHodlingPublicKeyResponse](nodeClient, routes.RoutePathIsHodlingPublicKey, *request, false, privateKey, "", isDerived)
}

func (nodeClient *NodeClient) GetAccessGroupMemberInfo(request *routes.GetAccessGroupMemberRequest, privateKey *btcec.PrivateKey, isDerived bool) (*routes.AccessGroupMemberEntryResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.GetAccessGroupMemberRequest, routes.AccessGroupMemberEntryResponse](nodeClient, routes.RoutePathGetAccessGroupMemberInfo, *request, false, privateKey, "", isDerived)
}

func (nodeClient *NodeClient) SendDiamonds(
	request *routes.SendDiamondsRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
) (*routes.SendDiamondsResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.SendDiamondsRequest, routes.SendDiamondsResponse](nodeClient, routes.RoutePathSendDiamonds, *request, true, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) CreateFollowTxn(
	request *routes.CreateFollowTxnStatelessRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
) (*routes.CreateFollowTxnStatelessResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.CreateFollowTxnStatelessRequest, routes.CreateFollowTxnStatelessResponse](nodeClient, routes.RoutePathCreateFollowTxnStateless, *request, true, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) SendMessage(
	request *routes.SendNewMessageRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
) (*routes.SendNewMessageResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.SendNewMessageRequest, routes.SendNewMessageResponse](nodeClient, routes.RoutePathSendDmMessage, *request, true, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) DAOCoins(
	request *routes.DAOCoinRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (*routes.DAOCoinResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.DAOCoinRequest, routes.DAOCoinResponse](nodeClient, routes.RoutePathDAOCoin, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) TransferDAOCoins(
	request *routes.TransferDAOCoinRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (*routes.TransferDAOCoinResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.TransferDAOCoinRequest, routes.TransferDAOCoinResponse](nodeClient, routes.RoutePathTransferDAOCoin, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) LockCoins(
	request *routes.CoinLockupRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (*routes.CoinLockResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.CoinLockupRequest, routes.CoinLockResponse](nodeClient, routes.RoutePathCoinLockup, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) UpdateGlobalParams(
	request *routes.UpdateGlobalParamsRequest,
	privateKey *btcec.PrivateKey,
	isDerived bool,
	signAndSubmitTxn bool,
) (*routes.UpdateGlobalParamsResponse, *routes.SubmitTransactionResponse, error) {
	return NodePostRequest[routes.UpdateGlobalParamsRequest, routes.UpdateGlobalParamsResponse](nodeClient, routes.RoutePathUpdateGlobalParams, *request, signAndSubmitTxn, privateKey, "TransactionHex", isDerived)
}

func (nodeClient *NodeClient) CheckPartyAccessGroups(request *routes.CheckPartyAccessGroupsRequest) (
	*routes.CheckPartyAccessGroupsResponse, error) {
	body, err := MakePostRequest(nodeClient.NodeURL, routes.RoutePathCheckPartyAccessGroups, request)
	if err != nil {
		return nil, fmt.Errorf("Error hitting CheckPartyAccessGroups: %v", err)
	}

	response := &routes.CheckPartyAccessGroupsResponse{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err = decoder.Decode(response); err != nil {
		return nil, fmt.Errorf(
			"Error parsing JSON response: %v", err)
	}

	return response, nil
}

func (nodeClient *NodeClient) AddAccessGroupMembers(
	request *routes.AddAccessGroupMembersRequest, derivedPublicKeyBase58Check string) (
	*routes.AddAccessGroupMembersResponse, error) {

	body, err := MakePostRequest(nodeClient.NodeURL, routes.RoutePathAddAccessGroupMembers, request)
	if err != nil {
		return nil, fmt.Errorf("Error hitting SendDeso: %v", err)
	}

	response := &routes.AddAccessGroupMembersResponse{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err = decoder.Decode(response); err != nil {
		return nil, fmt.Errorf(
			"Error parsing JSON response: %v", err)
	}

	if _, err = nodeClient.SignAndSubmitTransaction(
		nil, response.TransactionHex, false); err != nil {

		return nil, fmt.Errorf("Error signing and submitting transaction: %v", err)
	}

	return response, nil
}

func (nodeClient *NodeClient) FetchSingleDerivedKey(ownerPublicKey string, derivedPublicKey string) (*routes.GetSingleDerivedKeyResponse, error) {
	url := fmt.Sprintf("%s%s/%s/%s", nodeClient.NodeURL, routes.RoutePathGetSingleDerivedKey, ownerPublicKey, derivedPublicKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Error building request: ")
	}

	req.Header.Set("Origin", nodeClient.NodeURL)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Error fetching single derived key: ")
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status returned "+"from %v: %v, %v", url, resp.Status, string(body))
	}

	response := &routes.GetSingleDerivedKeyResponse{}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(response); err != nil {
		return nil, errors.Wrapf(err, "Error parsing JSON response:")
	}

	return response, nil
}

func MakePostRequest(nodeURL string, endpoint string, request interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", nodeURL, endpoint)
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Origin", nodeURL)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error Status returned "+
			"from %v: %v, %v", url, resp.Status, string(body))
	}
	return body, nil
}

func (nodeClient *NodeClient) MakeGetRequest(endpoint string, userPublicKeyBase58Check string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", nodeClient.NodeURL, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Origin", nodeClient.NodeURL)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error Status returned "+
			"from %v: %v, %v", url, resp.Status, string(body))
	}
	return body, nil
}

func (nodeClient *NodeClient) GetPublicKeyForPkid(pkid string) (string, error) {
	// Retrieve the public key for the user's pkid.
	profile := &entries.PGPkidEntry{}
	err := nodeClient.StateSyncerDB.NewSelect().Model(profile).Column("public_key").Where("pkid = ?", pkid).Scan(context.Background())
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("PayDueSubscription: Problem getting profile: %v", err)
	}

	if err == sql.ErrNoRows {
		return pkid, nil
	}
	return profile.PublicKey, nil
}

func (nodeClient *NodeClient) GetPkidForPublicKey(publicKey string) (string, error) {
	// Retrieve the public key for the user's pkid.
	profile := &entries.PGPkidEntry{}
	err := nodeClient.StateSyncerDB.NewSelect().Model(profile).Column("pkid").Where("public_key = ?", publicKey).Scan(context.Background())
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("PayDueSubscription: Problem getting profile: %v", err)
	}

	if err == sql.ErrNoRows {
		return publicKey, nil
	}
	return profile.Pkid, nil
}

// Search for transaction in postgres database, continue polling until transaction is returned.
func (nodeClient *NodeClient) WaitForTxnHash(txnHash string, isMined bool) (*entries.PGTransactionEntry, error) {

	var txnRes = &entries.PGTransactionEntry{}

	var err error

	fmt.Printf("Here is the txn hash we are waiting for: %s\n", txnHash)

	for txnRes == nil || err == sql.ErrNoRows || txnRes.TransactionHash != txnHash || (isMined && txnRes.BlockHeight == 0) {
		err = nodeClient.StateSyncerDB.NewSelect().Column("transaction_hash", "block_height").Model(txnRes).Where("transaction_hash = ?", txnHash).Scan(context.Background())
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("Error fetching transaction: %v", err)
		}
		time.Sleep(time.Millisecond * 10)
	}

	return txnRes, nil
}
