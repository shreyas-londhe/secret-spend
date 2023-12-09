package db

import (
	"math/big"
	"sync"

	"github.com/shreyas-londhe/private-erc20-circuits/merkletree"
	"github.com/shreyas-londhe/private-erc20-circuits/paillier"
)

type UserData struct {
	Index      int
	KeyPair    *paillier.PrivateKey
	Balance    *big.Int
	EncBalance *big.Int
	EncR       *big.Int
}

type UserResponse struct {
	Index      int                 `json:"index"`
	KeyPair    *paillier.PublicKey `json:"keyPair"`
	Balance    string              `json:"balance"`
	EncBalance string              `json:"encBalance"`
	EncR       string              `json:"encR"`
}

type DB struct {
	sync.RWMutex
	Users      []UserData
	MerkleTree *merkletree.MerkleTree
}

func New() *DB {
	return &DB{
		Users: make([]UserData, 0),
	}
}

func (db *DB) StoreUser(user UserData) {
	db.Lock()
	defer db.Unlock()
	db.Users = append(db.Users, user)
}

func (db *DB) StoreUserAtIndex(user UserData, index int) {
	db.Lock()
	defer db.Unlock()
	db.Users[index] = user
}

func (db *DB) StoreMerkleTree(tree *merkletree.MerkleTree) {
	db.Lock()
	db.MerkleTree = tree
	db.Unlock()
}

func (db *DB) GetAllUsers() []UserData {
	db.RLock()
	defer db.RUnlock()

	usersCopy := make([]UserData, len(db.Users))
	copy(usersCopy, db.Users)

	return usersCopy
}

func (db *DB) GetUser(index int) UserData {
	db.RLock()
	defer db.RUnlock()

	return db.Users[index]
}

func (db *DB) GetMerkleTree() *merkletree.MerkleTree {
	db.RLock()
	tree := db.MerkleTree
	db.RUnlock()
	return tree
}

func (db *DB) GetMerkleRoot() []byte {
	db.RLock()
	tree := db.MerkleTree
	db.RUnlock()
	return tree.MerkleRoot()
}

func (db *DB) GetMerkleProof(index int) ([][]byte, big.Int, error) {
	db.RLock()
	tree := db.MerkleTree
	db.RUnlock()
	leaf := db.Users[index]
	content := convertToLeaf(leaf)
	proof, proofHelper, err := tree.GetMerklePath(content)
	if err != nil {
		return nil, big.Int{}, err
	}

	success, err := tree.VerifyContent(content)
	if err != nil {
		return nil, big.Int{}, err
	}
	if !success {
		return nil, big.Int{}, err
	}

	return proof, proofHelper, nil
}
