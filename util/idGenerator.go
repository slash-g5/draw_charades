package util

import (
	"crypto/rand"
	"fmt"
	"time"
)

func GenerateGuid() string {
	timestamp := fmt.Sprintf("%x", time.Now().UnixNano())
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	randomStr := fmt.Sprintf("%x", randomBytes)

	return timestamp + randomStr
}
