package main

import (
	"sync"
)

var racer map[int]int

var race sync.RWMutex

func RacyReader() {
	for k, v := range racer {
		race.RLock() // Lock after read
		_, _ = k, v
		race.RUnlock()
	}
}

func SafeReader() {
	race.RLock() // Lock map
	for k, v := range racer {
		_, _ = k, v
	}
	race.RUnlock()
}

func Write() {
	for i := 0; i < 1e7; i++ {
		race.Lock()
		racer[i/2] = i
		race.Unlock()
	}
}

func main() {
	racer = make(map[int]int)
	Write()
	go Write()
	go Write()
	RacyReader()
	// SafeReader()
}
