package circuits

import (
	"github.com/consensys/gnark/frontend"
	gHash "github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/shreyas-londhe/private-erc20-circuits/utils"
)

type BalanceLeaf struct {
	PubKey     PaillierPubKey
	EncBalance frontend.Variable
}

type PrivateCoinCircuit struct {
	// Public inputs
	OldBalancesRoot frontend.Variable `gnark:",public"`
	NewBalancesRoot frontend.Variable `gnark:",public"`
	OldFromLeaf     BalanceLeaf       `gnark:",public"`
	OldToLeaf       BalanceLeaf       `gnark:",public"`
	NewFromLeaf     BalanceLeaf       `gnark:",public"`
	NewToLeaf       BalanceLeaf       `gnark:",public"`

	// Private inputs
	OldFromLeafMP       utils.MerkleProof
	OldFromLeafMPHelper frontend.Variable
	OldToLeafMP         utils.MerkleProof
	OldToLeafMPHelper   frontend.Variable
	OldFromBalance      frontend.Variable
	EncOldFromBalanceR  frontend.Variable
	EncNewFromBalanceR  frontend.Variable
	Amount              frontend.Variable
	EncAmountR          frontend.Variable
	NewFromLeafMP       utils.MerkleProof
	NewFromLeafMPHelper frontend.Variable
	NewToLeafMP         utils.MerkleProof
	NewToLeafMPHelper   frontend.Variable
}

func verifyMerkleProof(api frontend.API, hFunc gHash.FieldHasher, leaf BalanceLeaf, root frontend.Variable, proof utils.MerkleProof, helper frontend.Variable) {
	oldFromLeaf := utils.HashInCircuit(hFunc, leaf.PubKey.N, leaf.PubKey.G, leaf.EncBalance)
	api.AssertIsEqual(proof.Path[0], oldFromLeaf)
	proof.VerifyProof(api, hFunc, helper)
	api.AssertIsEqual(root, proof.RootHash)
}

func (circuit *PrivateCoinCircuit) Define(api frontend.API) error {
	hFunc, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	verifyMerkleProof(api, &hFunc, circuit.OldFromLeaf, circuit.OldBalancesRoot, circuit.OldFromLeafMP, circuit.OldFromLeafMPHelper)
	verifyMerkleProof(api, &hFunc, circuit.OldToLeaf, circuit.OldBalancesRoot, circuit.OldToLeafMP, circuit.OldToLeafMPHelper)

	encBal := circuit.OldFromLeaf.PubKey.Encrypt(api, circuit.OldFromBalance, circuit.EncOldFromBalanceR)
	api.AssertIsEqual(encBal, circuit.OldFromLeaf.EncBalance)

	api.AssertIsLessOrEqual(circuit.Amount, circuit.OldFromBalance)

	newFromBalance := api.Sub(circuit.OldFromBalance, circuit.Amount)
	encNewFromBalance := circuit.OldFromLeaf.PubKey.Encrypt(api, newFromBalance, circuit.EncNewFromBalanceR)
	api.AssertIsEqual(encNewFromBalance, circuit.NewFromLeaf.EncBalance)

	encAmount := circuit.OldToLeaf.PubKey.Encrypt(api, circuit.Amount, circuit.EncAmountR)
	newToLeafEncBalance := circuit.OldToLeaf.PubKey.Add(api, circuit.OldToLeaf.EncBalance, encAmount)
	api.AssertIsEqual(newToLeafEncBalance, circuit.NewToLeaf.EncBalance)

	verifyMerkleProof(api, &hFunc, circuit.NewFromLeaf, circuit.NewBalancesRoot, circuit.NewFromLeafMP, circuit.NewFromLeafMPHelper)
	verifyMerkleProof(api, &hFunc, circuit.NewToLeaf, circuit.NewBalancesRoot, circuit.NewToLeafMP, circuit.NewToLeafMPHelper)

	circuit.OldFromLeaf.PubKey.AssertIsEqual(api, circuit.NewFromLeaf.PubKey)
	circuit.OldToLeaf.PubKey.AssertIsEqual(api, circuit.NewToLeaf.PubKey)

	return nil
}
