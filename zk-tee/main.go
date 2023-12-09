package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/shreyas-londhe/private-erc20-circuits/db"
)

var database *db.DB

const (
	depth    int = 5
	numUsers int = 1 << depth
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Allow only your client app to access
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")   // Allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")         // Allowed headers

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Assuming you have a global variable 'database' of type *db.DB
	users := database.GetAllUsers()

	var response []db.UserResponse
	for _, user := range users {
		response = append(response, db.UserResponse{
			KeyPair: &user.KeyPair.PublicKey, // Send only the public part
			Balance: user.Balance.String(),
			Index:   user.Index,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getMerkleRootHandler(w http.ResponseWriter, r *http.Request) {
	tree := database.GetMerkleTree()
	if tree == nil {
		http.Error(w, "Merkle tree not found", http.StatusNotFound)
		return
	}

	type response struct {
		MerkleRoot []byte `json:"merkleRoot"`
	}

	resp := response{
		MerkleRoot: tree.MerkleRoot(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func transferFundsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	fromIndex, err := strconv.Atoi(r.URL.Query().Get("fromIndex"))
	if err != nil {
		http.Error(w, "Invalid fromIndex", http.StatusBadRequest)
		return
	}

	toIndex, err := strconv.Atoi(r.URL.Query().Get("toIndex"))
	if err != nil {
		http.Error(w, "Invalid toIndex", http.StatusBadRequest)
		return
	}

	if fromIndex == toIndex {
		http.Error(w, "fromIndex and toIndex cannot be the same", http.StatusBadRequest)
		return
	}

	amountStr := r.URL.Query().Get("amount")
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	tree := database.GetMerkleTree()
	users := database.GetAllUsers()

	witness, pInputs, fromUser, toUser, newTree, err := db.GenerateTransferWitness(depth, *tree, users, fromIndex, toIndex, amount)
	if err != nil {
		http.Error(w, "Error generating witness: "+err.Error(), http.StatusInternalServerError)
		return
	}
	database.StoreUserAtIndex(fromUser, fromIndex)
	database.StoreUserAtIndex(toUser, toIndex)
	database.StoreMerkleTree(&newTree)

	err = db.GenerateProofFromTransferWitness(witness, pInputs, depth)
	if err != nil {
		http.Error(w, "Error generating proof: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Assuming the proof JSON is saved in "exports/proof_data.json"
	proofJSON, err := os.ReadFile("exports/proof_data.json")
	if err != nil {
		http.Error(w, "Error reading proof file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the proof JSON as the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(proofJSON)
}

func main() {
	router := http.NewServeMux()

	handlerWithCors := corsMiddleware(router)

	database = db.New()

	users := db.GenerateData(numUsers)
	for _, user := range users {
		database.StoreUser(user)
	}

	tree := db.GenerateTreeFromUserData(users)
	database.StoreMerkleTree(&tree)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		allUsers := database.GetAllUsers()
		for _, user := range allUsers {
			fmt.Fprintf(w, "User Index: %d, Balance: %s\n", user.Index, user.Balance.String())
		}
	})

	router.HandleFunc("/get-all-users", getAllUsersHandler)

	router.HandleFunc("/get-merkle-root", getMerkleRootHandler)

	router.HandleFunc("/transfer-funds", transferFundsHandler)

	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", handlerWithCors); err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
