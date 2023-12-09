package utils

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
)

const (
	PaillierBits = 62
)

func RandomBigInt(n int) *big.Int {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return new(big.Int).Rand(rng, new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(n)), nil))
}

func HashInCircuit(h hash.FieldHasher, inputs ...frontend.Variable) frontend.Variable {
	h.Reset()
	for _, input := range inputs {
		h.Write(input)
	}
	res := h.Sum()

	return res
}

func Pad32Bytes(input []byte) []byte {
	inputLen := len(input)

	if inputLen < 32 {
		pad := make([]byte, 32-inputLen)
		return append(pad, input...)
	} else if inputLen > 32 {
		padSize := 32 - (inputLen % 32)
		if padSize == 32 {
			return input
		}
		pad := make([]byte, padSize)
		return append(pad, input...)
	}

	return input
}

func BitsToBigInt(bits []int64) *big.Int {
	result := new(big.Int)
	for i, bit := range bits {
		if bit == 1 {
			result.SetBit(result, len(bits)-i-1, 1)
		}
	}
	return result
}
