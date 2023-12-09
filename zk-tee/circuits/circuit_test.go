package circuits

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/shreyas-londhe/private-erc20-circuits/merkletree"
	"github.com/shreyas-londhe/private-erc20-circuits/paillier"
	"github.com/shreyas-londhe/private-erc20-circuits/utils"
)

type UserData struct {
	PubKey     paillier.PublicKey
	Balance    *big.Int
	EncBalance *big.Int
	EncR       *big.Int
}

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
	hfunc.Reset()
	hfunc.Write(utils.Pad32Bytes(t.PubKey.N.Bytes()))
	hfunc.Write(utils.Pad32Bytes(t.PubKey.G.Bytes()))
	hfunc.Write(utils.Pad32Bytes(t.EncBalance.Bytes()))
	return hfunc.Sum(nil), nil
}

func (t TestBalanceLeaf) Equals(other merkletree.Content) (bool, error) {
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

func GenerateRandomTree(depth int) (merkletree.MerkleTree, []TestBalanceLeaf, []UserData) {
	numLeaves := 1 << depth

	var data []UserData

	// Creating data
	keypairs := make([]*paillier.PrivateKey, numLeaves)
	encryptedBalances := make([]*big.Int, numLeaves)
	for i := 0; i < numLeaves; i++ {
		var err error
		keypairs[i], err = paillier.GenerateKey(rand.Reader, utils.PaillierBits)
		if err != nil {
			panic(err)
		}

		balanceBytes := utils.RandomBigInt(utils.PaillierBits - 1).Bytes()
		encryptedBalanceBytes, r, err := paillier.Encrypt(&keypairs[i].PublicKey, balanceBytes)
		if err != nil {
			panic(err)
		}
		balance := new(big.Int).SetBytes(balanceBytes)
		encryptedBalances[i] = new(big.Int).SetBytes(encryptedBalanceBytes)

		data = append(data, UserData{
			PubKey:     keypairs[i].PublicKey,
			Balance:    balance,
			EncBalance: encryptedBalances[i],
			EncR:       r,
		})
	}

	// Creating leaves
	var leavesInTree []merkletree.Content
	var leaves []TestBalanceLeaf
	for i := 0; i < numLeaves; i++ {
		leavesInTree = append(leavesInTree, TestBalanceLeaf{
			PubKey: TestPaillierPubKey{
				N: keypairs[i].PublicKey.N,
				G: keypairs[i].PublicKey.G,
			},
			EncBalance: encryptedBalances[i],
		})
		leaves = append(leaves, TestBalanceLeaf{
			PubKey: TestPaillierPubKey{
				N: keypairs[i].PublicKey.N,
				G: keypairs[i].PublicKey.G,
			},
			EncBalance: encryptedBalances[i],
		})
	}

	tree, err := merkletree.NewTree(leavesInTree)
	if err != nil {
		panic(err)
	}

	return *tree, leaves, data
}

func GenerateTreeFromLeaves(leaves []TestBalanceLeaf) merkletree.MerkleTree {
	var leavesInTree []merkletree.Content
	for _, leaf := range leaves {
		leavesInTree = append(leavesInTree, TestBalanceLeaf{
			PubKey: TestPaillierPubKey{
				N: leaf.PubKey.N,
				G: leaf.PubKey.G,
			},
			EncBalance: leaf.EncBalance,
		})
	}

	tree, err := merkletree.NewTree(leavesInTree)
	if err != nil {
		panic(err)
	}

	return *tree
}

func TestMainCircuit(t *testing.T) {
	assert := test.NewAssert(t)

	depth := 5

	testCase := func() {
		// Generate random tree
		tree, leaves, data := GenerateRandomTree(depth)

		var circuit PrivateCoinCircuit
		circuit.OldFromLeafMP.Path = make([]frontend.Variable, depth+1)
		circuit.OldToLeafMP.Path = make([]frontend.Variable, depth+1)
		circuit.NewFromLeafMP.Path = make([]frontend.Variable, depth+1)
		circuit.NewToLeafMP.Path = make([]frontend.Variable, depth+1)

		// Generate witness
		var witness PrivateCoinCircuit
		witness.OldFromLeafMP.Path = make([]frontend.Variable, depth+1)
		witness.OldToLeafMP.Path = make([]frontend.Variable, depth+1)
		witness.NewFromLeafMP.Path = make([]frontend.Variable, depth+1)
		witness.NewToLeafMP.Path = make([]frontend.Variable, depth+1)

		witness.OldBalancesRoot = tree.MerkleRoot()

		// For leaf 0
		proof0, proofHelper0, err := tree.GetMerklePath(leaves[0])
		if err != nil {
			panic(err)
		}
		success, err := tree.VerifyContent(leaves[0])
		if err != nil {
			panic(err)
		}
		assert.True(success)

		witness.OldFromLeaf.PubKey.N = leaves[0].PubKey.N
		witness.OldFromLeaf.PubKey.G = leaves[0].PubKey.G
		witness.OldFromLeaf.EncBalance = leaves[0].EncBalance
		witness.OldFromLeafMP.RootHash = tree.MerkleRoot()
		for i := 0; i < depth+1; i++ {
			if i == 0 {
				witness.OldFromLeafMP.Path[i], err = leaves[0].CalculateHash()
				if err != nil {
					panic(err)
				}
				continue
			}
			witness.OldFromLeafMP.Path[i] = proof0[i-1]
		}
		witness.OldFromLeafMPHelper = proofHelper0
		witness.OldFromBalance = data[0].Balance
		witness.EncOldFromBalanceR = data[0].EncR

		// For leaf 1
		proof1, proofHelper1, err := tree.GetMerklePath(leaves[1])
		if err != nil {
			panic(err)
		}
		success, err = tree.VerifyContent(leaves[1])
		if err != nil {
			panic(err)
		}
		assert.True(success)

		witness.OldToLeaf.PubKey.N = leaves[1].PubKey.N
		witness.OldToLeaf.PubKey.G = leaves[1].PubKey.G
		witness.OldToLeaf.EncBalance = leaves[1].EncBalance
		witness.OldToLeafMP.RootHash = tree.MerkleRoot()
		for i := 0; i < depth+1; i++ {
			if i == 0 {
				witness.OldToLeafMP.Path[i], err = leaves[1].CalculateHash()
				if err != nil {
					panic(err)
				}
				continue
			}
			witness.OldToLeafMP.Path[i] = proof1[i-1]
		}
		witness.OldToLeafMPHelper = proofHelper1

		// Encrypt amount with leaf 1's public key
		amount := big.NewInt(100) // Amount to transfer
		witness.Amount = amount
		encAmountBytes, r, err := paillier.Encrypt(&data[1].PubKey, amount.Bytes())
		if err != nil {
			panic(err)
		}
		witness.EncAmountR = r

		// Calculate new balance for leaf 0
		newFromBalance := new(big.Int).Sub(data[0].Balance, amount)
		encNewFromBalanceBytes, r, err := paillier.Encrypt(&data[0].PubKey, newFromBalance.Bytes())
		if err != nil {
			panic(err)
		}
		witness.EncNewFromBalanceR = r

		data[0].Balance = newFromBalance
		data[0].EncBalance = new(big.Int).SetBytes(encNewFromBalanceBytes)
		leaves[0].EncBalance = new(big.Int).SetBytes(encNewFromBalanceBytes)

		// Calculate new balance for leaf 1
		encNewToBalanceBytes := paillier.AddCipher(&data[1].PubKey, encAmountBytes, data[1].EncBalance.Bytes())
		data[1].Balance = new(big.Int).Add(amount, data[1].Balance)
		data[1].EncBalance = new(big.Int).SetBytes(encNewToBalanceBytes)
		leaves[1].EncBalance = new(big.Int).SetBytes(encNewToBalanceBytes)

		// Recalculate Tree
		newTree := GenerateTreeFromLeaves(leaves)
		witness.NewBalancesRoot = newTree.MerkleRoot()

		// For leaf 0
		newProof0, newProofHelper0, err := newTree.GetMerklePath(leaves[0])
		if err != nil {
			panic(err)
		}
		success, err = newTree.VerifyContent(leaves[0])
		if err != nil {
			panic(err)
		}
		assert.True(success)

		witness.NewFromLeaf.PubKey.N = leaves[0].PubKey.N
		witness.NewFromLeaf.PubKey.G = leaves[0].PubKey.G
		witness.NewFromLeaf.EncBalance = leaves[0].EncBalance
		witness.NewFromLeafMP.RootHash = newTree.MerkleRoot()
		for i := 0; i < depth+1; i++ {
			if i == 0 {
				witness.NewFromLeafMP.Path[i], err = leaves[0].CalculateHash()
				if err != nil {
					panic(err)
				}
				continue
			}
			witness.NewFromLeafMP.Path[i] = newProof0[i-1]
		}
		witness.NewFromLeafMPHelper = newProofHelper0

		// For leaf 1
		newProof1, newProofHelper1, err := newTree.GetMerklePath(leaves[1])
		if err != nil {
			panic(err)
		}
		success, err = newTree.VerifyContent(leaves[1])
		if err != nil {
			panic(err)
		}
		assert.True(success)

		witness.NewToLeaf.PubKey.N = leaves[1].PubKey.N
		witness.NewToLeaf.PubKey.G = leaves[1].PubKey.G
		witness.NewToLeaf.EncBalance = leaves[1].EncBalance
		witness.NewToLeafMP.RootHash = newTree.MerkleRoot()
		for i := 0; i < depth+1; i++ {
			if i == 0 {
				witness.NewToLeafMP.Path[i], err = leaves[1].CalculateHash()
				if err != nil {
					panic(err)
				}
				continue
			}
			witness.NewToLeafMP.Path[i] = newProof1[i-1]
		}
		witness.NewToLeafMPHelper = newProofHelper1

		err = test.IsSolved(&circuit, &witness, ecc.BN254.ScalarField())
		assert.NoError(err)
	}

	testCase()
}
