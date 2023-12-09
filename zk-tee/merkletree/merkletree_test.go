package merkletree

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/hash"
	"github.com/shreyas-londhe/private-erc20-circuits/utils"
)

type TestPaillierPubKey struct {
	N *big.Int
	G *big.Int
}

type TestBalanceLeaf struct {
	PubKey     TestPaillierPubKey
	EncBalance *big.Int
}

func (t TestBalanceLeaf) CalculateHash() ([]byte, error) {
	hfunc := hash.MIMC_BN254.New()
	hfunc.Write(utils.Pad32Bytes(t.PubKey.N.Bytes()))
	hfunc.Write(utils.Pad32Bytes(t.PubKey.G.Bytes()))
	hfunc.Write(utils.Pad32Bytes(t.EncBalance.Bytes()))
	return hfunc.Sum(nil), nil
}

func (t TestBalanceLeaf) Equals(other Content) (bool, error) {
	tHash, err := t.CalculateHash()
	if err != nil {
		return false, err
	}
	otherHash, err := other.CalculateHash()
	if err != nil {
		return false, err
	}
	return bytes.Equal(tHash, otherHash), nil
}

// TestMerkleTreeWithBalanceLeaf tests the Merkle tree functionality using BalanceLeaf.
func TestMerkleTreeWithBalanceLeaf(t *testing.T) {
	// Create a list of BalanceLeaf for testing
	var list []Content
	list = append(list, TestBalanceLeaf{
		PubKey:     TestPaillierPubKey{N: big.NewInt(123), G: big.NewInt(456)},
		EncBalance: big.NewInt(1000),
	})
	list = append(list, TestBalanceLeaf{
		PubKey:     TestPaillierPubKey{N: big.NewInt(789), G: big.NewInt(1011)},
		EncBalance: big.NewInt(2000),
	})

	// Create a new Merkle Tree from the list of Content
	tree, err := NewTree(list)
	if err != nil {
		t.Errorf("Failed to create Merkle Tree: %s", err)
	}

	// Test Merkle Root calculation
	mr := tree.MerkleRoot()
	if mr == nil {
		t.Error("Merkle Root should not be nil")
	}

	// Test tree verification
	validTree, err := tree.VerifyTree()
	if err != nil {
		t.Errorf("Failed to verify Merkle Tree: %s", err)
	}
	if !validTree {
		t.Error("Merkle Tree is invalid, but it should be valid")
	}

	// Test content verification
	for _, content := range list {
		validContent, err := tree.VerifyContent(content)
		if err != nil {
			t.Errorf("Failed to verify content: %s", err)
		}
		if !validContent {
			t.Error("Content is invalid, but it should be valid")
		}
	}
}
