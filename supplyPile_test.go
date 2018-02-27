package goburnbooks_test

import (
	gbb "goburnbooks"
	"sort"
	"strconv"
	"testing"
	"time"
)

func newRandomSupplyPile(count int, offset int) gbb.FSupplyPile {
	books := make([]gbb.Suppliable, count)

	for i := 0; i < count; i++ {
		id := offset + i
		params := &gbb.BookParams{BurnDuration: 0, ID: strconv.Itoa(id)}
		books[i] = gbb.NewBook(params)
	}

	bookParams := &gbb.SupplyPileParams{Supply: books}
	return gbb.NewSupplyPile(bookParams)
}

func Test_SupplyTakersHavingOddCapacity_ShouldStillLoadAll(t *testing.T) {
	/// Setup
	t.Parallel()
	supplyPiles := make([]gbb.FSupplyPile, supplyPileCount)
	t.Logf("Have %d supplies in total", totalSupplyCount)

	for ix := range supplyPiles {
		pile := newRandomSupplyPile(supplyPerPileCount, ix*supplyPerPileCount)
		supplyPiles[ix] = pile
	}

	pileGroup := gbb.NewSupplyPileGroup(supplyPiles...)
	takerCount := 50
	supplyTakers := make([]gbb.SupplyTaker, takerCount)
	supplyChs := make([]chan []gbb.Suppliable, takerCount)

	for ix := range supplyTakers {
		loadSupplyCh := make(chan []gbb.Suppliable)
		readyCh := make(chan interface{})

		btRawParams := &gbb.SupplyTakerRawParams{
			Cap:  13,
			STID: strconv.Itoa(ix),
		}

		btParams := &gbb.SupplyTakerParams{
			SupplyTakerRawParams: btRawParams,
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

		supply := gbb.NewSupplyTaker(btParams)
		supplyTakers[ix] = supply
		supplyChs[ix] = loadSupplyCh
	}

	// This ensures that loaded supplies are unique.
	loadedSupplies := make(map[gbb.Suppliable]int, 0)
	updateLoaded := make(chan []gbb.Suppliable)

	for _, ch := range supplyChs {
		go func(ch chan []gbb.Suppliable) {
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
		go func(taker gbb.SupplyTaker) {
			pileGroup.Supply(taker)
		}(taker)
	}

	time.Sleep(waitDuration)

	/// Then
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
		t.Logf("Supply taker %s took %d supplies", key, value)

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
