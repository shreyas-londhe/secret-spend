package paillier

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/shreyas-londhe/private-erc20-circuits/utils"
)

func TestPaillierCryptosystem(t *testing.T) {
	// Generate a n-bit private key.
	privKey, err := GenerateKey(rand.Reader, utils.PaillierBits)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Encrypt the number 15.
	m15 := new(big.Int).SetInt64(15)
	c15, _, err := Encrypt(&privKey.PublicKey, m15.Bytes())
	if err != nil {
		t.Fatalf("Failed to encrypt 15: %v", err)
	}

	// Decrypt the number 15.
	d, err := Decrypt(privKey, c15)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	plainText := new(big.Int).SetBytes(d)
	if plainText.Int64() != 15 {
		t.Errorf("Decryption of 15 failed, got %s", plainText.String())
	}

	// Encrypt the number 20.
	m20 := new(big.Int).SetInt64(20)
	c20, _, err := Encrypt(&privKey.PublicKey, m20.Bytes())
	if err != nil {
		t.Fatalf("Failed to encrypt 20: %v", err)
	}

	// Add the encrypted integers 15 and 20 together.
	plusM15M20 := AddCipher(&privKey.PublicKey, c15, c20)
	decryptedAddition, err := Decrypt(privKey, plusM15M20)
	if err != nil {
		t.Fatalf("Failed to decrypt addition: %v", err)
	}
	if new(big.Int).SetBytes(decryptedAddition).Int64() != 35 {
		t.Errorf("Addition of 15+20 failed, got %s", new(big.Int).SetBytes(decryptedAddition).String())
	}

	// Add the encrypted integer 15 to plaintext constant 10.
	plusE15and10 := Add(&privKey.PublicKey, c15, new(big.Int).SetInt64(10).Bytes())
	decryptedAddition, err = Decrypt(privKey, plusE15and10)
	if err != nil {
		t.Fatalf("Failed to decrypt addition with constant: %v", err)
	}
	if new(big.Int).SetBytes(decryptedAddition).Int64() != 25 {
		t.Errorf("Addition of 15+10 failed, got %s", new(big.Int).SetBytes(decryptedAddition).String())
	}

	// Multiply the encrypted integer 15 by the plaintext constant 10.
	mulE15and10 := Mul(&privKey.PublicKey, c15, new(big.Int).SetInt64(10).Bytes())
	decryptedMul, err := Decrypt(privKey, mulE15and10)
	if err != nil {
		t.Fatalf("Failed to decrypt multiplication: %v", err)
	}
	if new(big.Int).SetBytes(decryptedMul).Int64() != 150 {
		t.Errorf("Multiplication of 15*10 failed, got %s", new(big.Int).SetBytes(decryptedMul).String())
	}
}
