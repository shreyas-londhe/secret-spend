package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/shreyas-londhe/private-erc20-circuits/hints"
	"github.com/shreyas-londhe/private-erc20-circuits/utils"
)

type PaillierPubKey struct {
	N frontend.Variable
	G frontend.Variable
}

func DivMod(api frontend.API, num frontend.Variable, mod frontend.Variable) (frontend.Variable, frontend.Variable) {
	res, err := api.NewHint(hints.DivModHint, 2, num, mod)
	if err != nil {
		panic(err)
	}

	api.AssertIsEqual(num, api.Add(api.Mul(res[0], mod), res[1]))

	return res[0], res[1]
}

func MulMod(api frontend.API, num1 frontend.Variable, num2 frontend.Variable, mod frontend.Variable) frontend.Variable {
	prod := api.Mul(num1, num2)
	_, prodMod := DivMod(api, prod, mod)

	return prodMod
}

func SquareMod(api frontend.API, num frontend.Variable, mod frontend.Variable) frontend.Variable {
	sqr := api.Mul(num, num)
	_, sqrMod := DivMod(api, sqr, mod)

	return sqrMod
}

func PowMod(api frontend.API, base frontend.Variable, exp frontend.Variable, mod frontend.Variable) frontend.Variable {
	expBits := api.ToBinary(exp, utils.PaillierBits)
	res := frontend.Variable(1)

	_, baseMod := DivMod(api, base, mod)

	for i := utils.PaillierBits - 1; i >= 0; i-- {
		expBit := expBits[i]

		_, res = DivMod(api, api.Mul(res, res), mod)

		_, temp := DivMod(api, api.Mul(res, baseMod), mod)

		res = api.Select(expBit, temp, res)
	}

	return res
}

func (p PaillierPubKey) Encrypt(api frontend.API, message frontend.Variable, r frontend.Variable) frontend.Variable {
	n_2 := api.Mul(p.N, p.N)

	g_m := PowMod(api, p.G, message, n_2)
	r_n := PowMod(api, r, p.N, n_2)
	c := MulMod(api, g_m, r_n, n_2)

	return c
}

func (p PaillierPubKey) Add(api frontend.API, c1 frontend.Variable, c2 frontend.Variable) frontend.Variable {
	n_2 := api.Mul(p.N, p.N)
	c := MulMod(api, c1, c2, n_2)

	return c
}

func (p PaillierPubKey) AssertIsEqual(api frontend.API, other PaillierPubKey) {
	api.AssertIsEqual(p.N, other.N)
	api.AssertIsEqual(p.G, other.G)
}
