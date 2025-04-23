package bloom

import (
	"math/rand"
	"testing"
)

// --- Benchmarks ---

var (
	benchData = make([]byte, 64)
	H1, H2    uint64
)

func init() {
	// Remplir benchData avec des octets pseudo-aléatoires
	r := rand.New(rand.NewSource(42))
	r.Read(benchData)
}

func BenchmarkSum64_XXHash(b *testing.B) {
	h := XXHasher{}
	b.SetBytes(int64(len(benchData)))
	for b.Loop() {
		H1, H2 = h.Sum64(benchData)
	}
}

func BenchmarkSum64m_Murmur3(b *testing.B) {
	h := MumurHasher{}
	b.SetBytes(int64(len(benchData)))
	for b.Loop() {
		H1, H2 = h.Sum64(benchData)
	}
}

func BenchmarkSum64xm_XXHashMurmur(b *testing.B) {
	h := MixedHasher{}
	b.SetBytes(int64(len(benchData)))
	for b.Loop() {
		H1, H2 = h.Sum64(benchData)
	}
}

func BenchmarkSum64hm_MapHash(b *testing.B) {
	h := MapHasher{}
	b.SetBytes(int64(len(benchData)))
	for b.Loop() {
		H1, H2 = h.Sum64(benchData)
	}
}
