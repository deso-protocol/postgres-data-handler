package tests

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/deso-protocol/core/lib"
	"github.com/tyler-smith/go-bip39"
	"math/big"
	"regexp"
)

// Encrypt encrypts data for the target public key using AES-128-GCM
func Encrypt(pubKey *btcec.PublicKey, msg []byte) ([]byte, error) {
	var pt bytes.Buffer

	ephemeral, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	pt.Write(ephemeral.PubKey().SerializeUncompressed())

	ecdhKey := btcec.GenerateSharedSecret(ephemeral, pubKey)
	hashedSecret := sha256.Sum256(ecdhKey)
	encryptionKey := hashedSecret[:16]

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	pt.Write(nonce)

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, msg, nil)

	tag := ciphertext[len(ciphertext)-gcm.NonceSize():]
	pt.Write(tag)
	ciphertext = ciphertext[:len(ciphertext)-len(tag)]
	pt.Write(ciphertext)

	return pt.Bytes(), nil
}

// Decrypt decrypts a passed message with a receiver private key, returns plaintext or decryption error
func Decrypt(privkey *btcec.PrivateKey, msg []byte) ([]byte, error) {
	// Message cannot be less than length of public key (65) + nonce (16) + tag (16)
	if len(msg) <= (1 + 32 + 32 + 16 + 16) {
		return nil, fmt.Errorf("invalid length of message")
	}

	pb := new(big.Int).SetBytes(msg[:65]).Bytes()
	pubKey, err := btcec.ParsePubKey(pb)
	if err != nil {
		return nil, err
	}

	ecdhKey := btcec.GenerateSharedSecret(privkey, pubKey)
	hashedSecret := sha256.Sum256(ecdhKey)
	encryptionKey := hashedSecret[:16]

	msg = msg[65:]
	nonce := msg[:16]
	tag := msg[16:32]

	ciphertext := bytes.Join([][]byte{msg[32:], tag}, nil)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create new aes block: %w", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcm cipher: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt ciphertext: %w", err)
	}

	return plaintext, nil
}

func ComputeKeysFromSeedWithNet(seedBytes []byte, index uint32, desoParams *lib.DeSoParams) (_pubKey *btcec.PublicKey, _privKey *btcec.PrivateKey, _btcAddress string, _err error) {
	// Get the pubkey and privkey from the seed. We use the Bitcoin parameters
	// to generate them.
	// TODO: We should get this from the DeSoParams, not reference them directly.
	netParams := desoParams.BitcoinBtcdParams

	masterKey, err := hdkeychain.NewMaster(seedBytes, netParams)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'masterKey' from seed (%v)", err)
	}

	// We follow BIP44 to generate the addresses. Recall it follows the following
	// semantic hierarchy:
	// * purpose' / coin_type' / account' / change / address_index
	// For the derivation path we use: m/44'/0'/0'/0/0. Recall that 0' means we're
	// computing a "hardened" key, which means the private key is present, and
	// that 0 (no apostrophe) means we're computing an "unhardened" key which means
	// the private key is not present.
	//
	// m/44'/0'/0'/0/0 also maps to the first
	// address you'd get if you put the user's seed into most standard
	// Bitcoin wallets (Mycelium, Electrum, Ledger, iancoleman, etc...).
	purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'purpose' from seed (%v)", err)
	}
	coinTypeKey, err := purpose.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'coinType' from seed (%v)", err)
	}
	accountKey, err := coinTypeKey.Derive(hdkeychain.HardenedKeyStart + index)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'accountKey' from seed (%v)", err)
	}
	changeKey, err := accountKey.Derive(0)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'changeKey' from seed (%v)", err)
	}
	addressKey, err := changeKey.Derive(0)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'addressKey' from seed (%v)", err)
	}

	pubKey, err := addressKey.ECPubKey()
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'pubKey' from seed (%v)", err)
	}
	privKey, err := addressKey.ECPrivKey()
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'privKey' from seed (%v)", err)
	}
	addressObj, err := addressKey.Address(netParams)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ComputeKeyFromSeed: Error encountered generating 'addressObj' from seed (%v)", err)
	}
	btcDepositAddress := addressObj.EncodeAddress()

	return pubKey, privKey, btcDepositAddress, nil
}

func PubKeyToBase58(publicKey *btcec.PublicKey, desoParams *lib.DeSoParams) string {
	return lib.Base58CheckEncode(publicKey.SerializeCompressed(), false, desoParams)
}

func PrivKeyToHex(privateKey *btcec.PrivateKey) string {
	return hex.EncodeToString(privateKey.Serialize())
}

func Base58ToPubKey(publicKeyBase58 string, desoParams *lib.DeSoParams) (*btcec.PublicKey, error) {
	publicKeyBytes, _, err := lib.Base58CheckDecode(publicKeyBase58)
	if err != nil {
		return nil, err
	}
	publicKey, err := btcec.ParsePubKey(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

func StandardDerivation(seed string, groupName string, account uint32, desoParams *lib.DeSoParams) (*btcec.PublicKey, *btcec.PrivateKey, error) {
	seedBytes, err := bip39.NewSeedWithErrorChecking(seed, groupName)
	if err != nil {
		return nil, nil, err
	}
	publicKey, privateKey, _, err := ComputeKeysFromSeedWithNet(seedBytes, account, desoParams)
	if err != nil {
		return nil, nil, err
	}
	return publicKey, privateKey, nil
}

func CreateRandomKeyPair(params *lib.DeSoParams) (*btcec.PublicKey, *btcec.PrivateKey, error) {
	randomPrivateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	return randomPrivateKey.PubKey(), randomPrivateKey, nil
}

// Function to decrypt the encrypted access group private key hex
func DecryptEncryptedPrivKeyHex(encryptedPrivKeyHex string, privKey *btcec.PrivateKey) (*btcec.PrivateKey, error) {
	// Convert the hex string back to bytes
	encryptedBytes, err := hex.DecodeString(encryptedPrivKeyHex)
	if err != nil {
		return nil, fmt.Errorf("Error decoding hex string: %v", err)
	}

	// Convert *btcec.PrivateKey to *ecdsa.PrivateKey
	ecdsaPrivKey := privKey.ToECDSA()

	// Decrypt the bytes
	decryptedBytes, err := lib.DecryptBytesWithPrivateKey(encryptedBytes, ecdsaPrivKey)
	if err != nil {
		return nil, fmt.Errorf("Error decrypting bytes: %v", err)
	}

	seedHex := string(decryptedBytes)
	privateKeyBytes, err := hex.DecodeString(seedHex)

	if err != nil {
		return nil, fmt.Errorf("Error decoding hex string: %v", err)
	}

	// Convert the decrypted bytes back to btcec.PrivateKey
	decryptedPrivateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	return decryptedPrivateKey, nil
}

func GetAuthorizeDerivedKeyMetadataWithTransactionSpendingLimit(
	ownerPrivateKey *btcec.PrivateKey,
	derivedPrivateKey *btcec.PrivateKey,
	expirationBlock uint64,
	transactionSpendingLimit *lib.TransactionSpendingLimit,
	isDeleted bool,
	blockHeight uint64) (*lib.AuthorizeDerivedKeyMetadata, error) {

	// Generate public key
	derivedPublicKey := derivedPrivateKey.PubKey().SerializeCompressed()

	// Determine operation type
	var operationType lib.AuthorizeDerivedKeyOperationType
	if isDeleted {
		operationType = lib.AuthorizeDerivedKeyOperationNotValid
	} else {
		operationType = lib.AuthorizeDerivedKeyOperationValid
	}

	// We randomly use standard or the metamask derived key access signature.
	var accessBytes []byte

	// Create access signature
	expirationBlockByte := lib.EncodeUint64(expirationBlock)
	accessBytes = append(derivedPublicKey, expirationBlockByte[:]...)

	var transactionSpendingLimitBytes []byte
	transactionSpendingLimitBytes, err := transactionSpendingLimit.ToBytes(blockHeight)
	accessBytes = append(accessBytes, transactionSpendingLimitBytes[:]...)
	if err != nil {
		return nil, err
	}

	accessSignature := ecdsa.Sign(ownerPrivateKey, lib.Sha256DoubleHash(accessBytes)[:]).Serialize()

	return &lib.AuthorizeDerivedKeyMetadata{
		derivedPublicKey,
		expirationBlock,
		operationType,
		accessSignature,
	}, nil
}

func DeriveDefaultMessagingKey(privKey *btcec.PrivateKey, messagingKeyName string) (*btcec.PrivateKey, *btcec.PublicKey) {
	seedHexBytes := privKey.Serialize()
	messagingPrivateKey := lib.Sha256DoubleHash(append(lib.Sha256DoubleHash(seedHexBytes)[:], lib.Sha256DoubleHash([]byte(messagingKeyName))[:]...))[:]
	return btcec.PrivKeyFromBytes(messagingPrivateKey)
}

// Convert a seed hex to a btcec.PrivateKey
func SeedHexToKeyPair(seedHex string) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	seedBytes, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, nil, err
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(seedBytes)

	return privKey, pubKey, nil
}

func SeedToKeyPair(seed string, password string, accountIndex uint32, desoParams *lib.DeSoParams) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	seedBytes, err := bip39.NewSeedWithErrorChecking(seed, password)
	if err != nil {
		return nil, nil, err
	}

	publicKey, privateKey, _, err := ComputeKeysFromSeedWithNet(seedBytes, accountIndex, desoParams)
	return privateKey, publicKey, err
}

func GetZeroPublicKeyBase58(desoParams *lib.DeSoParams) string {
	return lib.Base58CheckEncode(lib.ZeroPublicKey.ToBytes(), false, desoParams)
}

// IsValidPublicKey checks if the publicKey is alphanumeric.
func IsValidPublicKey(publicKey string) bool {
	// This regular expression matches strings that are only alphanumeric.
	re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return re.MatchString(publicKey)
}
