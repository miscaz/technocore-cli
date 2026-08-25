package technocore

import (
	"errors"
	"math/big"
)

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(input []byte) string {
	x := new(big.Int).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var out []byte
	for x.Cmp(zero) > 0 {
		x.DivMod(x, base, mod)
		out = append(out, base58Alphabet[mod.Int64()])
	}
	for _, b := range input {
		if b != 0 {
			break
		}
		out = append(out, base58Alphabet[0])
	}
	// reverse
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func base58Decode(s string) ([]byte, error) {
	result := big.NewInt(0)
	base := big.NewInt(58)
	for _, r := range s {
		idx := -1
		for i := 0; i < len(base58Alphabet); i++ {
			if base58Alphabet[i] == byte(r) {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, errors.New("invalid base58 character")
		}
		result.Mul(result, base)
		result.Add(result, big.NewInt(int64(idx)))
	}
	decoded := result.Bytes()
	// restore leading zeros
	var zeros int
	for _, r := range s {
		if byte(r) == base58Alphabet[0] {
			zeros++
		} else {
			break
		}
	}
	return append(make([]byte, zeros), decoded...), nil
}
