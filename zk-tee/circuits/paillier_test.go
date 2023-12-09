package circuits

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/shreyas-londhe/private-erc20-circuits/paillier"
	"github.com/shreyas-londhe/private-erc20-circuits/utils"
)

type TestDivModCircuit struct {
	Num       frontend.Variable
	Mod       frontend.Variable
	Quotient  frontend.Variable
	Remainder frontend.Variable
}

type TestPowModCircuit struct {
	Base   frontend.Variable
	Exp    frontend.Variable
	Mod    frontend.Variable
	Result frontend.Variable
}

type TestPaillierEncryptionCircuit struct {
	Message1  frontend.Variable
	Message2  frontend.Variable
	R1        frontend.Variable
	R2        frontend.Variable
	Cipher1   frontend.Variable
	Cipher2   frontend.Variable
	CipherSum frontend.Variable
	PubKey    PaillierPubKey
}

func (circuit *TestDivModCircuit) Define(api frontend.API) error {
	q, r := DivMod(api, circuit.Num, circuit.Mod)

	api.AssertIsEqual(q, circuit.Quotient)
	api.AssertIsEqual(r, circuit.Remainder)

	return nil
}

func (circuit *TestPowModCircuit) Define(api frontend.API) error {
	result := PowMod(api, circuit.Base, circuit.Exp, circuit.Mod)
	api.AssertIsEqual(result, circuit.Result)

	return nil
}

func (circuit *TestPaillierEncryptionCircuit) Define(api frontend.API) error {
	c1 := circuit.PubKey.Encrypt(api, circuit.Message1, circuit.R1)
	api.AssertIsEqual(c1, circuit.Cipher1)

	c2 := circuit.PubKey.Encrypt(api, circuit.Message2, circuit.R2)
	api.AssertIsEqual(c2, circuit.Cipher2)

	cSum := circuit.PubKey.Add(api, c1, c2)
	api.AssertIsEqual(cSum, circuit.CipherSum)

	return nil
}

func TestDivMod(t *testing.T) {
	assert := test.NewAssert(t)

	num := utils.RandomBigInt(utils.PaillierBits)
	mod := utils.RandomBigInt(utils.PaillierBits)
	quotient := new(big.Int).Quo(num, mod)
	remainder := new(big.Int).Rem(num, mod)

	testCase := func() {
		// Create a new DivModCircuit instance with test values
		circuit := &TestDivModCircuit{}
		witness := &TestDivModCircuit{
			Num:       num,
			Mod:       mod,
			Quotient:  quotient,
			Remainder: remainder,
		}

		err := test.IsSolved(circuit, witness, ecc.BN254.ScalarField())
		assert.NoError(err)
	}
	testCase()
}

func TestPowMod(t *testing.T) {
	assert := test.NewAssert(t)

	base := utils.RandomBigInt(utils.PaillierBits)
	exp := utils.RandomBigInt(utils.PaillierBits)
	mod := utils.RandomBigInt(utils.PaillierBits)
	result := new(big.Int).Exp(base, exp, mod)

	testCase := func() {
		// Create a new TestPowModCircuit instance with test values
		circuit := &TestPowModCircuit{}
		witness := &TestPowModCircuit{
			Base:   base,
			Exp:    exp,
			Mod:    mod,
			Result: result,
		}

		err := test.IsSolved(circuit, witness, ecc.BN254.ScalarField())
		assert.NoError(err)
	}
	testCase()
}

func TestPaillierEncryption(t *testing.T) {
	assert := test.NewAssert(t)

	privKey, err := paillier.GenerateKey(rand.Reader, utils.PaillierBits)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	n := privKey.PublicKey.N
	g := privKey.PublicKey.G
	pubKey := PaillierPubKey{
		N: n,
		G: g,
	}

	message1 := utils.RandomBigInt(utils.PaillierBits - 1)
	cipher1, r1, err := paillier.Encrypt(&privKey.PublicKey, message1.Bytes())
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	message2 := utils.RandomBigInt(utils.PaillierBits - 1)
	cipher2, r2, err := paillier.Encrypt(&privKey.PublicKey, message2.Bytes())
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	cipherSum := paillier.AddCipher(&privKey.PublicKey, cipher1, cipher2)

	testCase := func() {
		// Create a new TestPowModCircuit instance with test values
		circuit := &TestPaillierEncryptionCircuit{}
		witness := &TestPaillierEncryptionCircuit{
			Message1:  message1,
			Message2:  message2,
			R1:        r1,
			R2:        r2,
			Cipher1:   cipher1,
			Cipher2:   cipher2,
			CipherSum: cipherSum,
			PubKey:    pubKey,
		}

		err := test.IsSolved(circuit, witness, ecc.BN254.ScalarField())
		assert.NoError(err)
	}
	testCase()
}
