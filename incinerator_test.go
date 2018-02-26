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
	t.Parallel()
	pending := make(chan Burner, capacity)
	incinerator := NewIncinerator(capacity, pending)
	ig := NewIncineratorGroup(incinerator)

	// When
	for i := 0; i < count; i++ {
		go func(ix int) {
			burner := NewBurner(strconv.Itoa(ix), defaultDuration)
			ig.Feed(burner)
		}(i)
	}

	// Then
	time.Sleep(waitDuration)
	length := len(ig.Burned())

	if length != count {
		t.Errorf("Should have burned %d, but instead got %d", count, length)
	}
}

func Test_BurnMultipleBurners_ShouldCapAtSpecifiedCapacity(t *testing.T) {
	// Setup
	t.Parallel()
	pending := make(chan Burner, capacity*1000)
	incinerator := NewIncinerator(capacity, pending)
	ig := NewIncineratorGroup(incinerator)

	// When
	for i := 0; i < count; i++ {
		go func(ix int) {
			// Unrealistic burn duration to simulate blocking operation.
			burner := NewBurner(strconv.Itoa(ix), 1e10)
			ig.Feed(burner)
		}(i)
	}

	// Then
	time.Sleep(waitDuration)
	length := len(ig.Burned())

	if length != 0 {
		t.Errorf("Should not have burned anything, but instead got %d", length)
	}
}
