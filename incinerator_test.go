package goburnbooks

import (
	"strconv"
	"testing"
	"time"
)

const (
	capacity        = 10
	count           = 1000
	defaultDuration = time.Duration(1)
	waitDuration    = 1e9
)

func Test_BurnMultipleBurners_ShouldEventuallyBurnAll(t *testing.T) {
	// Setup
	pending := make(chan Burner, capacity)
	incinerator := NewIncinerator(capacity, pending)

	// When
	for i := 0; i < count; i++ {
		go func(ix int) {
			burner := NewBurner(strconv.Itoa(ix), defaultDuration)
			pending <- burner
		}(i)
	}

	// Then
	allBurned := make([]Burner, 0)

	go func() {
		for burned := range incinerator.Burned() {
			allBurned = append(allBurned, burned)
		}
	}()

	time.Sleep(waitDuration)
	length := len(allBurned)

	if length != count {
		t.Errorf("Should have burned %d, but instead got %d", count, length)
	}
}
