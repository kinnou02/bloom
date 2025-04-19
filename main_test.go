package main

import (
	"blo/bloom"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	mainFilterPath = "bloomfilter.dat"
	defaultN       = 1_000_000_000
)

func BenchmarkBloomFromMainFile(b *testing.B) {
	if _, err := os.Stat(mainFilterPath); os.IsNotExist(err) {
		b.Fatalf("File %s not found. Run `go run main.go` first to generate it.", mainFilterPath)
	}

	filter, cleanup, err := bloom.LoadFromFile(mainFilterPath)
	require.NoError(b, err)
	defer cleanup()
	nbFound := 0.0
	count := 0.0
	b.ResetTimer()
	for b.Loop() {
		count++
		key := []byte(fmt.Sprintf("key-%d", rand.Intn(defaultN)))
		found := filter.Test(key)
		if found {
			nbFound++
		}
	}
	log.Printf("found %f on %f (%f)", nbFound, count, (nbFound / count * 100))
}

func BenchmarkBloomParallelTest(b *testing.B) {
	if _, err := os.Stat(mainFilterPath); os.IsNotExist(err) {
		b.Fatalf("File %s not found. Run `go run main.go` first to generate it.", mainFilterPath)
	}

	filter, cleanup, err := bloom.LoadFromFile(mainFilterPath)
	require.NoError(b, err)
	defer cleanup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(42))
		for pb.Next() {
			key := []byte(fmt.Sprintf("key-%d", r.Intn(defaultN)))
			filter.Test(key)
		}
	})
}
