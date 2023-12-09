package db

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/shreyas-londhe/private-erc20-circuits/circuits"
	"github.com/shreyas-londhe/private-erc20-circuits/hints"
	"github.com/shreyas-londhe/private-erc20-circuits/merkletree"
	"github.com/shreyas-londhe/private-erc20-circuits/paillier"
	"github.com/shreyas-londhe/private-erc20-circuits/utils"
)

type BalanceLeaf struct {
	PubKey     PaillierPubKey
	EncBalance *big.Int
}

type PaillierPubKey struct {
	N *big.Int
	G *big.Int
}

type Groth16ProofData struct {
	Proof  []string `json:"proof"`
	Inputs []string `json:"inputs"`
}

func (t BalanceLeaf) CalculateHash() ([]byte, error) {
	hfunc := hash.MIMC_BN254.New()
	hfunc.Reset()
	hfunc.Write(utils.Pad32Bytes(t.PubKey.N.Bytes()))
	hfunc.Write(utils.Pad32Bytes(t.PubKey.G.Bytes()))
	hfunc.Write(utils.Pad32Bytes(t.EncBalance.Bytes()))
	return hfunc.Sum(nil), nil
}

func (t BalanceLeaf) Equals(other merkletree.Content) (bool, error) {
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

func convertToLeaf(user UserData) BalanceLeaf {
	return BalanceLeaf{
		PubKey: PaillierPubKey{
			N: user.KeyPair.PublicKey.N,
			G: user.KeyPair.PublicKey.G,
		},
		EncBalance: user.EncBalance,
	}
}

func GenerateData(n int) []UserData {
	var users []UserData
	for i := 0; i < n; i++ {
		keyPair, err := paillier.GenerateKey(rand.Reader, utils.PaillierBits)
		if err != nil {
			panic(err)
		}

		balance := utils.RandomBigInt(utils.PaillierBits - 1)
		encBalance, r, err := paillier.Encrypt(&keyPair.PublicKey, balance.Bytes())
		if err != nil {
			panic(err)
		}

		user := UserData{
			Index:      i,
			KeyPair:    keyPair,
			Balance:    balance,
			EncBalance: new(big.Int).SetBytes(encBalance),
			EncR:       r,
		}
		users = append(users, user)
	}
	return users
}

func GenerateTreeFromUserData(users []UserData) merkletree.MerkleTree {
	var leaves []merkletree.Content
	for _, user := range users {
		leaf := convertToLeaf(user)
		leaves = append(leaves, leaf)
	}

	tree, err := merkletree.NewTree(leaves)
	if err != nil {
		panic(err)
	}

	return *tree
}

func GenerateTransferWitness(
	depth int,
	tree merkletree.MerkleTree,
	users []UserData,
	fromIndex int,
	toIndex int,
	amount *big.Int,
) (circuits.PrivateCoinCircuit, [14]*big.Int, UserData, UserData, merkletree.MerkleTree, error) {
	var pubInputs [14]*big.Int

	var circuit circuits.PrivateCoinCircuit
	circuit.OldFromLeafMP.Path = make([]frontend.Variable, depth+1)
	circuit.OldToLeafMP.Path = make([]frontend.Variable, depth+1)
	circuit.NewFromLeafMP.Path = make([]frontend.Variable, depth+1)
	circuit.NewToLeafMP.Path = make([]frontend.Variable, depth+1)

	var witness circuits.PrivateCoinCircuit
	witness.OldFromLeafMP.Path = make([]frontend.Variable, depth+1)
	witness.OldToLeafMP.Path = make([]frontend.Variable, depth+1)
	witness.NewFromLeafMP.Path = make([]frontend.Variable, depth+1)
	witness.NewToLeafMP.Path = make([]frontend.Variable, depth+1)

	witness.OldBalancesRoot = tree.MerkleRoot()
	pubInputs[0] = new(big.Int).SetBytes(tree.MerkleRoot())

	// For leaf fromIndex
	leaf0 := users[fromIndex]
	content0 := convertToLeaf(leaf0)
	proof0, proofHelper0, err := tree.GetMerklePath(content0)
	if err != nil {
		panic(err)
	}
	success, err := tree.VerifyContent(content0)
	if err != nil {
		panic(err)
	}
	if !success {
		panic("failed to verify content")
	}

	witness.OldFromLeaf.PubKey.N = content0.PubKey.N
	pubInputs[2] = new(big.Int).SetBytes(content0.PubKey.N.Bytes())
	witness.OldFromLeaf.PubKey.G = content0.PubKey.G
	pubInputs[3] = new(big.Int).SetBytes(content0.PubKey.G.Bytes())
	witness.OldFromLeaf.EncBalance = content0.EncBalance
	pubInputs[4] = new(big.Int).SetBytes(content0.EncBalance.Bytes())
	witness.OldFromLeafMP.RootHash = tree.MerkleRoot()
	for i := 0; i < depth+1; i++ {
		if i == 0 {
			witness.OldFromLeafMP.Path[i], err = content0.CalculateHash()
			if err != nil {
				panic(err)
			}
			continue
		}
		witness.OldFromLeafMP.Path[i] = proof0[i-1]
	}
	witness.OldFromLeafMPHelper = proofHelper0
	witness.OldFromBalance = leaf0.Balance
	witness.EncOldFromBalanceR = leaf0.EncR

	// For leaf toIndex
	leaf1 := users[toIndex]
	content1 := convertToLeaf(leaf1)
	proof1, proofHelper1, err := tree.GetMerklePath(content1)
	if err != nil {
		panic(err)
	}
	success, err = tree.VerifyContent(content1)
	if err != nil {
		panic(err)
	}
	if !success {
		panic("failed to verify content")
	}

	witness.OldToLeaf.PubKey.N = content1.PubKey.N
	pubInputs[5] = new(big.Int).SetBytes(content1.PubKey.N.Bytes())
	witness.OldToLeaf.PubKey.G = content1.PubKey.G
	pubInputs[6] = new(big.Int).SetBytes(content1.PubKey.G.Bytes())
	witness.OldToLeaf.EncBalance = content1.EncBalance
	pubInputs[7] = new(big.Int).SetBytes(content1.EncBalance.Bytes())
	witness.OldToLeafMP.RootHash = tree.MerkleRoot()
	for i := 0; i < depth+1; i++ {
		if i == 0 {
			witness.OldToLeafMP.Path[i], err = content1.CalculateHash()
			if err != nil {
				panic(err)
			}
			continue
		}
		witness.OldToLeafMP.Path[i] = proof1[i-1]
	}
	witness.OldToLeafMPHelper = proofHelper1

	// For Amount
	witness.Amount = amount
	encAmountBytes, r, err := paillier.Encrypt(&leaf1.KeyPair.PublicKey, amount.Bytes())
	if err != nil {
		panic(err)
	}
	witness.EncAmountR = r

	// Calculate new balance for leaf fromIndex
	newFromBalance := new(big.Int).Sub(leaf0.Balance, amount)
	encNewFromBalanceBytes, r, err := paillier.Encrypt(&leaf0.KeyPair.PublicKey, newFromBalance.Bytes())
	if err != nil {
		panic(err)
	}
	witness.EncNewFromBalanceR = r

	leaf0.Balance = newFromBalance
	leaf0.EncBalance = new(big.Int).SetBytes(encNewFromBalanceBytes)
	leaf0.EncR = r
	content0 = convertToLeaf(leaf0)
	err = tree.ModifyLeafAt(fromIndex, content0)
	if err != nil {
		panic(err)
	}

	// Calculate new balance for leaf toIndex
	encNewToBalanceBytes := paillier.AddCipher(&leaf1.KeyPair.PublicKey, encAmountBytes, leaf1.EncBalance.Bytes())
	leaf1.Balance.Add(leaf1.Balance, amount)
	leaf1.EncBalance = new(big.Int).SetBytes(encNewToBalanceBytes)
	leaf1.EncR = r
	content1 = convertToLeaf(leaf1)
	err = tree.ModifyLeafAt(toIndex, content1)
	if err != nil {
		panic(err)
	}

	witness.NewBalancesRoot = tree.MerkleRoot()
	pubInputs[1] = new(big.Int).SetBytes(tree.MerkleRoot())

	// For leaf fromIndex after transfer
	newProof0, newProofHelper0, err := tree.GetMerklePath(content0)
	if err != nil {
		panic(err)
	}
	success, err = tree.VerifyContent(content0)
	if err != nil {
		panic(err)
	}
	if !success {
		panic("failed to verify content")
	}

	witness.NewFromLeaf.PubKey.N = content0.PubKey.N
	pubInputs[8] = new(big.Int).SetBytes(content0.PubKey.N.Bytes())
	witness.NewFromLeaf.PubKey.G = content0.PubKey.G
	pubInputs[9] = new(big.Int).SetBytes(content0.PubKey.G.Bytes())
	witness.NewFromLeaf.EncBalance = content0.EncBalance
	pubInputs[10] = new(big.Int).SetBytes(content0.EncBalance.Bytes())
	witness.NewFromLeafMP.RootHash = tree.MerkleRoot()
	for i := 0; i < depth+1; i++ {
		if i == 0 {
			witness.NewFromLeafMP.Path[i], err = content0.CalculateHash()
			if err != nil {
				panic(err)
			}
			continue
		}
		witness.NewFromLeafMP.Path[i] = newProof0[i-1]
	}
	witness.NewFromLeafMPHelper = newProofHelper0

	// For leaf toIndex after transfer
	newProof1, newProofHelper1, err := tree.GetMerklePath(content1)
	if err != nil {
		panic(err)
	}
	success, err = tree.VerifyContent(content1)
	if err != nil {
		panic(err)
	}
	if !success {
		panic("failed to verify content")
	}

	witness.NewToLeaf.PubKey.N = content1.PubKey.N
	pubInputs[11] = new(big.Int).SetBytes(content1.PubKey.N.Bytes())
	witness.NewToLeaf.PubKey.G = content1.PubKey.G
	pubInputs[12] = new(big.Int).SetBytes(content1.PubKey.G.Bytes())
	witness.NewToLeaf.EncBalance = content1.EncBalance
	pubInputs[13] = new(big.Int).SetBytes(content1.EncBalance.Bytes())
	witness.NewToLeafMP.RootHash = tree.MerkleRoot()
	for i := 0; i < depth+1; i++ {
		if i == 0 {
			witness.NewToLeafMP.Path[i], err = content1.CalculateHash()
			if err != nil {
				panic(err)
			}
			continue
		}
		witness.NewToLeafMP.Path[i] = newProof1[i-1]
	}
	witness.NewToLeafMPHelper = newProofHelper1

	return witness, pubInputs, leaf0, leaf1, tree, nil
}

func GenerateProofFromTransferWitness(witness circuits.PrivateCoinCircuit, pInputs [14]*big.Int, depth int) error {
	var circuit circuits.PrivateCoinCircuit
	circuit.OldFromLeafMP.Path = make([]frontend.Variable, depth+1)
	circuit.OldToLeafMP.Path = make([]frontend.Variable, depth+1)
	circuit.NewFromLeafMP.Path = make([]frontend.Variable, depth+1)
	circuit.NewToLeafMP.Path = make([]frontend.Variable, depth+1)

	CreateNewKeys := false
	var ccs constraint.ConstraintSystem
	var pk groth16.ProvingKey
	var vk groth16.VerifyingKey

	if CreateNewKeys {
		var err error

		ccs, err = frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
		if err != nil {
			return err
		}

		pk, vk, err = groth16.Setup(ccs)
		if err != nil {
			return err
		}

		{
			f, err := os.Create("exports/circuit.r1cs")
			if err != nil {
				return err
			}
			_, err = ccs.WriteTo(f)
			if err != nil {
				return err
			}
		}
		{
			f, err := os.Create("exports/circuit.vk")
			if err != nil {
				return err
			}
			_, err = vk.WriteRawTo(f)
			if err != nil {
				return err
			}
		}
		{
			f, err := os.Create("exports/circuit.pk")
			if err != nil {
				return err
			}
			_, err = pk.WriteRawTo(f)
			if err != nil {
				return err
			}
		}
		{
			f, err := os.Create("exports/verifier.sol")
			if err != nil {
				return err
			}
			err = vk.ExportSolidity(f)
			if err != nil {
				return err
			}
		}
		fmt.Println("Wrote keys to exports/circuit.pk and exports/circuit.vk")
	} else {
		ccs = groth16.NewCS(ecc.BN254)
		fmt.Println("Reading circuit from exports/circuit.r1cs")
		{
			f, err := os.Open("exports/circuit.r1cs")
			if err != nil {
				return err
			}
			_, err = ccs.ReadFrom(f)
			if err != nil {
				return err
			}
		}
		pk = groth16.NewProvingKey(ecc.BN254)
		fmt.Println("Reading proving key from exports/circuit.pk")
		{
			f, _ := os.Open("exports/circuit.pk")
			_, err := pk.ReadFrom(f)
			f.Close()
			if err != nil {
				return err
			}
		}
		vk = groth16.NewVerifyingKey(ecc.BN254)
		fmt.Println("Reading verifying key from exports/circuit.vk")
		{
			f, _ := os.Open("exports/circuit.vk")
			_, err := vk.ReadFrom(f)
			f.Close()
			if err != nil {
				return err
			}
		}
		{
			f, err := os.Create("exports/verifier.sol")
			if err != nil {
				return err
			}
			err = vk.ExportSolidity(f)
			if err != nil {
				return err
			}
		}
		println("Read keys from exports/circuit.pk and exports/circuit.vk")
	}

	validWitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		return err
	}

	validPublicWitness, err := validWitness.Public()
	if err != nil {
		return err
	}

	proof, err := groth16.Prove(ccs, pk, validWitness, backend.WithSolverOptions(solver.WithHints(hints.DivModHint)))
	if err != nil {
		return err
	}
	fmt.Println("Proving done.")

	err = groth16.Verify(proof, vk, validPublicWitness)
	if err != nil {
		return err
	}
	fmt.Println("Verifying done.")

	GenerateProofData(proof, pInputs, 14)

	return nil
}

func GenerateProofData(proof groth16.Proof, pubInputs [14]*big.Int, pubInputLen int) {
	const fpSize = 4 * 8
	var buf bytes.Buffer
	proof.WriteRawTo(&buf)
	proofBytes := buf.Bytes()

	proofs := make([]string, 8)
	for i := 0; i < 8; i++ {
		proofs[i] = "0x" + hex.EncodeToString(proofBytes[i*fpSize:(i+1)*fpSize])
	}

	inputs := make([]string, pubInputLen)
	for i := 0; i < pubInputLen; i++ {
		inputs[i] = "0x" + fmt.Sprintf("%x", pubInputs[i])
	}

	// Create the data struct and populate it
	data := Groth16ProofData{
		Proof:  proofs,
		Inputs: inputs,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	err = os.WriteFile("exports/proof_data.json", jsonData, 0o644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}
