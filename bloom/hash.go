package bloom

import (
	"github.com/cespare/xxhash/v2"
)

type Hasher struct{}

func NewHasher() *Hasher {
	return &Hasher{}
}

func (h *Hasher) Sum64(data []byte) (uint64, uint64) {
	h1 := xxhash.Sum64(data)

	salted := append(data, 'X') // simple dérivation
	h2 := xxhash.Sum64(salted)

	return h1, h2
}
