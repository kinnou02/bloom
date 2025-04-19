package main

import (
	"blo/bloom"
	"log"
	"net/http"
	"sync"
)

const filterPath = "bloomfilter.dat"

var (
	filter  *bloom.BloomFilter
	cleanup func()
	bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 256)
		},
	}
)

func Start() {
	var err error
	filter, cleanup, err = bloom.LoadFromFile(filterPath)
	if err != nil {
		log.Fatalf("failed to load bloom filter: %v", err)
	}
	defer cleanup()

	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing ?key=...", http.StatusBadRequest)
			return
		}

		buf := bufPool.Get().([]byte)
		buf = append(buf[:0], key...) // reset + reuse
		defer bufPool.Put(buf)

		if filter.Test(buf) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("1"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("0"))
		}
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
