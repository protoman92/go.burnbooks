package goburnbooks

import (
	"sort"
	"strconv"
	"testing"
	"time"
)

const (
	defaultTimeout = 1e8
)

func newRandomSupplyPile(count int, offset int) FSupplyPile {
	books := make([]Suppliable, count)

	for i := 0; i < count; i++ {
		id := offset + i
		params := &BookParams{BurnDuration: 0, ID: strconv.Itoa(id)}
		books[i] = NewBook(params)
	}

	bookParams := &SupplyPileParams{supply: books}
	return NewSupplyPile(bookParams)
}

func Test_SupplyTakersHavingOddCapacity_ShouldStillLoadAll(t *testing.T) {
	/// Setup
	t.Parallel()
	pileCount, supplyCount := 20, 10000
	totalBCount := pileCount * supplyCount
	supplyPiles := make([]FSupplyPile, pileCount)
	t.Logf("Have %d supplies in total", totalBCount)

	for ix := range supplyPiles {
		pile := newRandomSupplyPile(supplyCount, ix*supplyCount)
		supplyPiles[ix] = pile
	}

	pileGroup := NewSupplyPileGroup(supplyPiles...)
	takerCount := 50
	supplyTakers := make([]SupplyTaker, takerCount)
	supplyChs := make([]chan []Suppliable, takerCount)

	for ix := range supplyTakers {
		loadSupplyCh := make(chan []Suppliable)
		readyCh := make(chan interface{})

		btParams := &SupplyTakerParams{
			capacity: 17,
			id:       strconv.Itoa(ix),
			loadCh:   loadSupplyCh,
			readyCh:  readyCh,
		}

		// Assume that the supply taker takes repeatedly at a specified delay.
		go func() {
			for {
				time.Sleep(1e5)

				go func() {
					readyCh <- true
				}()
			}
		}()

		supply := NewSupplyTaker(btParams)
		supplyTakers[ix] = supply
		supplyChs[ix] = loadSupplyCh
	}

	// This ensures that loaded supplies are unique.
	loadedSupplies := make(map[Suppliable]int, 0)
	updateLoaded := make(chan []Suppliable)

	for _, ch := range supplyChs {
		go func(ch chan []Suppliable) {
			for {
				select {
				case burnables := <-ch:
					go func() {
						updateLoaded <- burnables
					}()
				}
			}
		}(ch)
	}

	go func() {
		for {
			select {
			case loaded := <-updateLoaded:
				if len(loaded) > 0 {
					for _, item := range loaded {
						loadedSupplies[item] = loadedSupplies[item] + 1
					}
				}
			}
		}
	}()

	/// When
	for _, taker := range supplyTakers {
		go func(taker SupplyTaker) {
			pileGroup.Supply(taker)
		}(taker)
	}

	time.Sleep(2e9)

	/// Then
	allTakenResult := pileGroup.Taken()
	allTakenMap := make(map[string]int, 0)
	allTakenCount := 0

	for _, result := range allTakenResult {
		takenCount := len(result.supplyIds)
		allTakenMap[result.takerID] = allTakenMap[result.takerID] + takenCount
		allTakenCount += takenCount
	}

	if allTakenCount != totalBCount {
		t.Errorf("Should have taken %d, but got %d", totalBCount, allTakenCount)
	}

	for key, value := range allTakenMap {
		t.Logf("Supply taker %s took %d supplies", key, value)

		if value == 0 {
			t.Errorf("%s should have taken some, but took nothing", key)
		}
	}

	loadedLen := len(loadedSupplies)

	if loadedLen != totalBCount {
		t.Errorf("Should have taken %d, instead got %d", totalBCount, loadedLen)
	}

	keys := make([]int, 0)

	for key, value := range loadedSupplies {
		numericKey, _ := strconv.Atoi(key.UID())
		keys = append(keys, numericKey)

		if value != 1 {
			t.Errorf("%v should have been taken once, but was %d", key, value)
		}
	}

	sort.Ints(keys)
	keyLen := len(keys)

	if keyLen != totalBCount {
		t.Errorf("Should have %d keys, instead got %d", totalBCount, keyLen)
	}
}
