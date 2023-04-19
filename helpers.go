package geploy

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
)

var CommandCounter uint32

func nextCommandId() uint32 { return atomic.AddUint32(&CommandCounter, +1) }

func randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}
