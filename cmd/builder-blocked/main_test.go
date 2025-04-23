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
	mainFilterPath = "../../blockedbloomfilter.dat"
	defaultN       = 1_000_000_000
)

func BenchmarkBloomFromMainFile(b *testing.B) {
	if _, err := os.Stat(mainFilterPath); os.IsNotExist(err) {
		b.Fatalf("File %s not found. Run `go run main.go` first to generate it.", mainFilterPath)
	}

	filter, err := bloom.LoadBlockedBloomFilter(mainFilterPath)
	require.NoError(b, err)
	defer filter.Close()
	nbFound := 0.0
	count := 0.0
	b.ResetTimer()
	FP := 0.0
	shouldBeKO := 0.0
	for b.Loop() {
		count++
		v := rand.Intn(defaultN)
		key := []byte(fmt.Sprintf("key-%d", v))
		found := filter.Test(key)
		if found {
			nbFound++
		}
		if v > 500_000_000 {
			shouldBeKO++
			if found {
				FP++
			}
		}
	}
	log.Printf("found %f on %f (%f)., fp: %f (fp rate: %f)", nbFound, count, (nbFound / count * 100), FP, FP/shouldBeKO)
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
