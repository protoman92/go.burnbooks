package goburnbooks

import (
	"strconv"
	"testing"
	"time"
)

const (
	capacity        = 500
	count           = 10000
	defaultDuration = 1e7
	minCapacity     = capacity / 2
	waitDuration    = 1e9
)

func Test_BurnMultiple_ShouldEventuallyBurnAll(t *testing.T) {
	/// Setup
	t.Parallel()
	pending := make(chan Burnable, capacity)

	iParams := IncineratorParams{
		Capacity:    capacity,
		MinCapacity: minCapacity,
		UID:         "1",
	}

	incinerator := NewIncinerator(iParams, pending)
	ig := NewIncineratorGroup(incinerator)

	/// When
	for i := 0; i < count; i++ {
		go func(ix int) {
			bParams := BookParams{BurnDuration: defaultDuration, UID: strconv.Itoa(ix)}
			burnable := NewBook(bParams)
			ig.Incinerate(burnable)
		}(i)
	}

	time.Sleep(waitDuration)

	/// Then
	length := len(ig.Burned())

	if length != count {
		t.Errorf("Should have burned %d, but instead got %d", count, length)
	}
}

func Test_BurnMultiple_ShouldCapAtSpecifiedCapacity(t *testing.T) {
	/// Setup
	t.Parallel()
	pending := make(chan Burnable, capacity*1000)

	iParams := IncineratorParams{
		Capacity:    capacity,
		MinCapacity: minCapacity,
		UID:         "1",
	}

	incinerator := NewIncinerator(iParams, pending)
	ig := NewIncineratorGroup(incinerator)

	/// When
	for i := 0; i < count; i++ {
		go func(ix int) {
			// Unrealistic burn duration to simulate blocking operation.
			bParams := BookParams{BurnDuration: 1e15, UID: strconv.Itoa(ix)}
			burnable := NewBook(bParams)
			ig.Incinerate(burnable)
		}(i)
	}

	time.Sleep(waitDuration)

	/// Then
	length := len(ig.Burned())

	if length != 0 {
		t.Errorf("Should not have burned anything, but instead got %d", length)
	}
}

func Test_BurnMultipleBooksWithIncineratorGroup_ShouldAllocate(t *testing.T) {
	/// Setup
	t.Parallel()
	oIncCount := 10
	id1 := "Blocker"
	otherIncs := make([]FIncinerator, oIncCount)

	for ix := range otherIncs {
		id := strconv.Itoa(ix)

		iParams := IncineratorParams{
			Capacity:    capacity,
			MinCapacity: minCapacity,
			UID:         id,
		}

		otherIncs[ix] = NewIncinerator(iParams, make(chan Burnable, capacity))
	}

	i1Params := IncineratorParams{UID: id1}
	i1 := NewIncinerator(i1Params, make(chan Burnable, capacity))
	foreverID := "Forever"

	// For the purpose of this test, this might as well be forever. Normally we
	// should not let individual incinerators directly handle the burning, and
	// instead delegate to an incinerator group for better resource allocation.
	foreverParams := BookParams{BurnDuration: 1e15, UID: foreverID}
	forever := NewBook(foreverParams)
	i1.Incinerate(forever)

	// Take out availability to prevent new Burnables from being added to pending
	// pile.
	<-i1.SignalAvailable()

	allIncs := append(otherIncs, i1)
	ig := NewIncineratorGroup(allIncs...)

	/// When
	for i := 0; i < count; i++ {
		go func(ix int) {
			bParams := BookParams{BurnDuration: defaultDuration, UID: strconv.Itoa(ix)}
			burnable := NewBook(bParams)
			ig.Incinerate(burnable)
		}(i)
	}

	time.Sleep(waitDuration)

	/// Then
	allBurned := ig.Burned()
	var i1Count, oCount int
	otherBurned := make(map[string]int, 0)

	for _, burned := range allBurned {
		incineratorID := burned.incineratorID

		switch incineratorID {
		case id1:
			i1Count++

		default:
			oCount++
			otherBurned[incineratorID] = otherBurned[incineratorID] + 1
		}
	}

	oBurnedLen := len(otherBurned)

	if oBurnedLen != oIncCount {
		t.Errorf("Expected %d incinerators, but got %d", oIncCount, oBurnedLen)
	}

	if i1Count != 0 {
		t.Errorf("i1 should not have burned anything, but instead got %d", i1Count)
	}

	for key, value := range otherBurned {
		t.Logf("Incinerator %s burned %d", key, value)

		if value == 0 {
			t.Errorf("%s should have burned something, but instead got nothing", key)
		}
	}

	if oCount != count {
		t.Errorf(
			"Other incinerators should have burned %d, but instead got %d",
			count,
			oCount,
		)
	}
}
