package main

import (
	"fmt"
	"math/rand"
	"time"

	"blo/bloom" // adapte ce chemin à ton module si besoin
)

const (
	numElements  = 500_000_000
	falsePosRate = 0.0001
	filterFile   = "blockedbloomfilter.dat"
)

func main() {
	filter, err := bloom.CreateBlockedBloomFilter(filterFile, numElements, falsePosRate)
	if err != nil {
		panic(err)
	}

	// Ajout d'éléments
	fmt.Println("Adding elements...")
	for i := 0; i < numElements; i++ {
		key := []byte(fmt.Sprintf("key-%d", i))
		filter.Add(key)
	}

	// Sauvegarde
	if err := filter.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("Saved Bloom filter to %s\n", filterFile)

	// Chargement mmap
	mapped, err := bloom.LoadBlockedBloomFilter(filterFile)
	if err != nil {
		panic(err)
	}
	defer mapped.Close()
	fmt.Println("Loaded Bloom filter from file (mmap)")

	// Vérif d'un élément connu
	testKey := []byte("key-42")
	fmt.Printf("Test key-42: %v (should be true)\n", mapped.Test(testKey))

	// Faux positifs
	fmt.Println("Testing false positives...")
	rand.Seed(time.Now().UnixNano())
	falsePos := 0
	total := 100_000
	for i := 0; i < total; i++ {
		k := fmt.Sprintf("zzz-%d", rand.Int())
		if mapped.Test([]byte(k)) {
			falsePos++
		}
	}
	fmt.Printf("False positives: %d / %d (%.4f)\n", falsePos, total, float64(falsePos)/float64(total))
}
