package testcases

import (
	"multiplayer_game/util"
	"testing"
)

func TestGenerateGuid(t *testing.T) {
	results := make(map[string]bool)
	for i := 0; i < 500000; i++ {
		curr := util.GenerateGuid()
		_, ok := results[curr]
		if ok {
			t.Fatalf("non unique values are generated in generateguid function")
		}
		results[curr] = true
	}
}
