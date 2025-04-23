package main

import (
	"blo/bloom"
	"log"
	"strings"
	"sync"

	"github.com/tidwall/redcon"
)

const (
	filterPath = "blockedbloomfilter.dat"
	addr       = ":6380" // port Redis habituel
)

var (
	filter  *bloom.BlockedBloomFilter
	bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 256)
		},
	}
)

func main() {
	var err error
	// Charge le bloom filter au démarrage
	filter, err = bloom.LoadBlockedBloomFilter(filterPath)
	if err != nil {
		log.Fatalf("failed to load bloom filter: %v", err)
	}
	defer filter.Close()

	log.Printf("Starting Redis‑like server on %s (BF.EXISTS only)\n", addr)
	err = redcon.ListenAndServe(addr,
		// Handle commande
		func(conn redcon.Conn, cmd redcon.Command) {
			name := strings.ToUpper(string(cmd.Args[0]))
			switch name {
			case "BF.EXISTS":
				// syntaxe : BF.EXISTS <key>
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for 'bf.exists' command")
					return
				}
				// Optionnel : réutiliser un buffer si besoin
				buf := bufPool.Get().([]byte)[:0]
				buf = append(buf, cmd.Args[2]...)
				defer bufPool.Put(buf)

				if filter.Test(buf) {
					conn.WriteInt(1)
				} else {
					conn.WriteInt(0)
				}
			default:
				conn.WriteError("ERR unknown command '" + name + "'")
			}
		},
		// On accepte une connexion
		func(conn redcon.Conn) bool {
			// return true pour accepter, false pour refuser
			return true
		},
		// On ferme une connexion
		func(conn redcon.Conn, err error) {
			// rien de particulier à faire
		},
	)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
