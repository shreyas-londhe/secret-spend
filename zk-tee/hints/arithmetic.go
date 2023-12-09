package hints

import "math/big"

func DivModHint(_ *big.Int, inputs, outputs []*big.Int) error {
	quotient := new(big.Int)
	remainder := new(big.Int)

	quotient.Quo(inputs[0], inputs[1])
	remainder.Rem(inputs[0], inputs[1])

	outputs[0].Set(quotient)
	outputs[1].Set(remainder)

	return nil
}
