package main

import (
	"blo/bloom"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	mainFilterPath = "../../blockedbloomfilter.dat"
	defaultN       = 1_000_000_000
)

var (
	found bool
	keys  [][]byte
)

func init() {
	keys = make([][]byte, 1_000_000)
	for i := 0; i < len(keys); i++ {
		keys[i] = []byte(fmt.Sprintf("key-%d", rand.Intn(defaultN)))
	}
}

func BenchmarkBloomFromMainFile(b *testing.B) {
	if _, err := os.Stat(mainFilterPath); os.IsNotExist(err) {
		b.Fatalf("File %s not found. Run `go run main.go` first to generate it.", mainFilterPath)
	}

	filter, err := bloom.LoadBlockedBloomFilter(mainFilterPath)
	require.NoError(b, err)
	defer filter.Close()
	count := 0
	fmt.Println("Starting benchmark...")
	b.ResetTimer()
	for b.Loop() {
		count++
		key := keys[count%len(keys)]
		found = filter.Test(key)
	}
}

func BenchmarkBloomParallelTest(b *testing.B) {
	if _, err := os.Stat(mainFilterPath); os.IsNotExist(err) {
		b.Fatalf("File %s not found. Run `go run main.go` first to generate it.", mainFilterPath)
	}

	filter, err := bloom.LoadBlockedBloomFilter(mainFilterPath)
	require.NoError(b, err)
	defer filter.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(42))
		for pb.Next() {
			key := []byte(fmt.Sprintf("key-%d", r.Intn(defaultN)))
			filter.Test(key)
		}
	})
}
