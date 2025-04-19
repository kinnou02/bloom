package main_test

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	stdbloom "github.com/bits-and-blooms/bloom/v3"
)

const (
	stdN          = 500_000_000
	stdFpRate     = 0.001
	stdFilterPath = "bitsetandbloom.dat"
)

func createAndSaveStdBloom(path string) (*stdbloom.BloomFilter, error) {
	filter := stdbloom.NewWithEstimates(uint(stdN), stdFpRate)
	for i := 0; i < stdN; i++ {
		filter.Add([]byte(fmt.Sprintf("key-%d", i)))
	}

	// Save to file using gob
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = gob.NewEncoder(f).Encode(filter)
	if err != nil {
		return nil, err
	}
	return filter, nil
}

func loadStdBloom(path string) (*stdbloom.BloomFilter, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var filter stdbloom.BloomFilter
	err = gob.NewDecoder(f).Decode(&filter)
	if err != nil {
		return nil, err
	}
	return &filter, nil
}

func BenchmarkBitsetAndBloomMemory(b *testing.B) {
	var filter *stdbloom.BloomFilter
	var err error

	if _, err = os.Stat(stdFilterPath); os.IsNotExist(err) {
		b.Logf("Fichier %s non trouvé, génération...", stdFilterPath)
		filter, err = createAndSaveStdBloom(stdFilterPath)
		if err != nil {
			b.Fatalf("Erreur génération filtre: %v", err)
		}
	} else {
		filter, err = loadStdBloom(stdFilterPath)
		if err != nil {
			b.Fatalf("Erreur chargement filtre: %v", err)
		}
	}

	nbFound := 0.0
	count := 0.0
	b.ResetTimer()
	for b.Loop() {
		count++
		key := []byte(fmt.Sprintf("key-%d", rand.Intn(1_000_000_000)))
		found := filter.Test(key)
		if found {
			nbFound++
		}
	}
	log.Printf("found %f on %f (%f)", nbFound, count, (nbFound / count * 100))
}
