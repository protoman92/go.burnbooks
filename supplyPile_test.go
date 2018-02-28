package goburnbooks

import (
	"sort"
	"strconv"
	"testing"
	"time"
)

func newRandomSupplyPile(count int, offset int) SupplyPile {
	books := make([]Suppliable, count)

	for i := 0; i < count; i++ {
		id := offset + i
		params := &BookParams{BurnDuration: 0, ID: strconv.Itoa(id)}
		books[i] = NewBook(params)
	}

	bookParams := &SupplyPileParams{
		Logger:      logWorker,
		Supply:      books,
		TakeTimeout: supplyPileTimeout,
	}

	return NewSupplyPile(bookParams)
}

func Test_SupplyTakersHavingOddCapacity_ShouldStillLoadAll(t *testing.T) {
	/// Setup
	t.Parallel()
	supplyPiles := make([]SupplyPile, supplyPileCount)

	for ix := range supplyPiles {
		pile := newRandomSupplyPile(supplyPerPileCount, ix*supplyPerPileCount)
		supplyPiles[ix] = pile
	}

	pileGroup := NewSupplyPileGroup(supplyPiles...)
	supplyTakers := make([]SupplyTaker, supplyTakerCount)
	supplyChs := make([]chan []Suppliable, supplyTakerCount)

	for ix := range supplyTakers {
		loadSupplyCh := make(chan []Suppliable)
		readyCh := make(chan interface{})

		stRawParams := &SupplyTakerRawParams{
			Cap:  13,
			STID: strconv.Itoa(ix),
		}

		stParams := &SupplyTakerParams{
			SupplyTakerRawParams: stRawParams,
			LoadCh:               loadSupplyCh,
			TakeReadyCh:          readyCh,
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

		supply := NewSupplyTaker(stParams)
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

	time.Sleep(waitDuration)

	/// Then
	verifySupplyGroupFairContrib(pileGroup, contribPercentThreshold, t)
	allTakenResults := pileGroup.Taken()
	allTakenMap := make(map[string]int, 0)
	allTakenCount := 0

	for _, result := range allTakenResults {
		takerID := result.TakerID
		takenCount := len(result.SupplyIds)
		allTakenMap[takerID] = allTakenMap[takerID] + takenCount
		allTakenCount += takenCount
	}

	if allTakenCount != totalSupplyCount {
		t.Errorf("Should have taken %d, but got %d", totalSupplyCount, allTakenCount)
	}

	for key, value := range allTakenMap {
		if value == 0 {
			t.Errorf("%s should have taken some, but took nothing", key)
		}
	}

	loadedSuppliesLen := len(loadedSupplies)

	if loadedSuppliesLen != totalSupplyCount {
		t.Errorf(
			"Should have taken %d, instead got %d",
			totalSupplyCount,
			loadedSuppliesLen,
		)
	}

	loadSupplyKeys := make([]int, 0)

	for key, value := range loadedSupplies {
		numericKey, _ := strconv.Atoi(key.SuppliableID())
		loadSupplyKeys = append(loadSupplyKeys, numericKey)

		if value != 1 {
			t.Errorf("%v should have been taken once, but was %d", key, value)
		}
	}

	sort.Ints(loadSupplyKeys)
	loadSupplyKeyLen := len(loadSupplyKeys)

	if loadSupplyKeyLen != totalSupplyCount {
		t.Errorf(
			"Should have %d keys, instead got %d",
			totalSupplyCount,
			loadSupplyKeyLen,
		)
	}
}
