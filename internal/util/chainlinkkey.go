package util

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

type Raw []byte

func (raw Raw) String() string {
	return "<CSA Raw Private Key>"
}

func (raw Raw) GoString() string {
	return raw.String()
}

func (raw Raw) Bytes() []byte {
	return ([]byte)(raw)
}

// EIP55Address is a new type for string which persists an ethereum address in
// its original string representation which includes a leading 0x, and EIP55
// checksum which is represented by the case of digits A-F.
type EIP55Address string

// NewEIP55Address creates an EIP55Address from a string, an error is returned if:
//
// 1) There is no leading 0x
// 2) The length is wrong
// 3) There are any non hexadecimal characters
// 4) The checksum fails
func NewEIP55Address(s string) (EIP55Address, error) {
	address := common.HexToAddress(s)
	if s != address.Hex() {
		return EIP55Address(""), fmt.Errorf(`"%s" is not a valid EIP55 formatted address`, s)
	}
	return EIP55Address(s), nil
}

// EIP55AddressFromAddress forces an address into EIP55Address format
// It is safe to panic on error since address.Hex() should ALWAYS generate EIP55Address-compatible hex strings
func EIP55AddressFromAddress(a common.Address) EIP55Address {
	addr, err := NewEIP55Address(a.Hex())
	if err != nil {
		panic(err)
	}
	return addr
}

type KeyV2 struct {
	Address      common.Address
	EIP55Address EIP55Address
	privateKey   *ecdsa.PrivateKey
}

func NewV2() (KeyV2, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return KeyV2{}, err
	}
	return FromPrivateKey(privateKeyECDSA), nil
}

func FromPrivateKey(privKey *ecdsa.PrivateKey) (key KeyV2) {
	address := crypto.PubkeyToAddress(privKey.PublicKey)
	eip55 := EIP55AddressFromAddress(address)
	return KeyV2{
		Address:      address,
		EIP55Address: eip55,
		privateKey:   privKey,
	}
}

func (key KeyV2) ID() string {
	return key.Address.Hex()
}

func (key KeyV2) ToEcdsaPrivKey() *ecdsa.PrivateKey {
	return key.privateKey
}

func (key KeyV2) Raw() Raw {
	return key.privateKey.D.Bytes()
}

func (key KeyV2) String() string {
	return fmt.Sprintf("EthKeyV2{PrivateKey: <redacted>, Address: %s}", key.Address)
}

func (key KeyV2) GoString() string {
	return key.String()
}

// Cmp uses byte-order address comparison to give a stable comparison between two keys
func (key KeyV2) Cmp(key2 KeyV2) int {
	return bytes.Compare(key.Address.Bytes(), key2.Address.Bytes())
}

func (key KeyV2) ToEncryptedJSON(password string, scryptParams ScryptParams) (export []byte, err error) {
	// DEV: uuid is derived directly from the address, since it is not stored internally
	id, err := uuid.FromBytes(key.Address.Bytes()[:16])
	if err != nil {
		return nil, fmt.Errorf("%w: could not generate ethkey UUID", err)
	}

	dKey := &keystore.Key{
		Id:         id,
		Address:    key.Address,
		PrivateKey: key.privateKey,
	}

	return keystore.EncryptKey(dKey, password, scryptParams.N, scryptParams.P)
}

const (
	// FastN is a shorter N parameter for testing
	FastN = 2
	// FastP is a shorter P parameter for testing
	FastP = 1
)

type (
	// ScryptParams represents two integers, N and P.
	ScryptParams struct{ N, P int }
	// ScryptConfigReader can check for an insecure, fast flag
	ScryptConfigReader interface {
		InsecureFastScrypt() bool
	}
)

// DefaultScryptParams is for use in production. It used geth's standard level
// of encryption and is relatively expensive to decode.
// Avoid using this in tests.
var DefaultScryptParams = ScryptParams{N: keystore.StandardScryptN, P: keystore.StandardScryptP}

// FastScryptParams is for use in tests, where you don't want to wear out your
// CPU with expensive key derivations, do not use it in production, or your
// encrypted keys will be easy to brute-force!
var FastScryptParams = ScryptParams{N: FastN, P: FastP}

// GetScryptParams fetches ScryptParams from a ScryptConfigReader
func GetScryptParams(config ScryptConfigReader) ScryptParams {
	if config.InsecureFastScrypt() {
		return FastScryptParams
	}
	return DefaultScryptParams
}
