package goburnbooks

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func Test_GopherDeliveringBurnables_ShouldBurnAll(t *testing.T) {
	/// Setup
	t.Parallel()
	gophers := make([]Gopher, gopherCount)

	for ix := range gophers {
		gParams := &GopherParams{
			BurnableProviderRawParams: &BurnableProviderRawParams{
				BPID: strconv.Itoa(ix),
			},
			SupplyTakerRawParams: &SupplyTakerRawParams{
				Cap:  gopherCapacity,
				STID: strconv.Itoa(ix),
			},
			Logger:       logWorker,
			TakeTimeout:  gopherTakeTimeout,
			TripDuration: tripDelay,
		}

		gopher := NewGopher(gParams)
		gophers[ix] = gopher
	}

	burnDuration := time.Duration(1)
	piles := make([]SupplyPile, supplyPileCount)
	allBooks := make([]Book, 0)
	allBookIds := make([]string, 0)

	for ix := range piles {
		supplies := make([]Suppliable, supplyPerPileCount)

		for jx := range supplies {
			id := fmt.Sprintf("%d-%d", ix, jx)
			bParams := &BookParams{BurnDuration: burnDuration, ID: id}
			book := NewBook(bParams)
			supplies[jx] = book
			allBooks = append(allBooks, book)
			allBookIds = append(allBookIds, id)
		}

		pParams := &SupplyPileParams{
			Logger:      logWorker,
			Supply:      supplies,
			ID:          strconv.Itoa(ix),
			TakeTimeout: supplyPileTimeout,
		}

		pile := NewSupplyPile(pParams)
		piles[ix] = pile
	}

	pileGroup := NewSupplyPileGroup(piles...)

	incinerators := make([]Incinerator, incineratorCount)

	for ix := range incinerators {
		iParams := &IncineratorParams{
			Capacity:    incineratorCap,
			ID:          strconv.Itoa(ix),
			Logger:      logWorker,
			MinCapacity: incineratorMinCap,
		}

		incinerator := NewIncinerator(iParams)
		incinerators[ix] = incinerator
	}

	incineratorGroup := NewIncineratorGroup(incinerators...)
	totalBookCount := len(allBookIds)

	/// When
	for _, gopher := range gophers {
		go pileGroup.Supply(gopher)
		go incineratorGroup.Consume(gopher)
	}

	time.Sleep(integrationWaitDuration)

	/// Then
	verifySupplyGroupFairContrib(pileGroup, contribPercentThreshold, t)
	verifyIncGroupFairContrib(incineratorGroup, contribPercentThreshold, t)
	allBookIDLen := len(allBookIds)

	if allBookIDLen != totalSupplyCount {
		t.Errorf("Should have %d books, but got %d", totalSupplyCount, allBookIDLen)
	}

	allBurned := incineratorGroup.Burned()
	allBurnedLen := len(allBurned)
	burnedIDMap := incineratorGroup.BurnedIDMap()
	burnedIDMapLen := len(burnedIDMap)
	incineratorMap := incineratorGroup.IncineratorContribMap()
	incineratorMapLen := len(incineratorMap)
	providerMap := incineratorGroup.ProviderContribMap()
	providerMapLen := len(providerMap)
	takenMap := pileGroup.SupplyTakerContribMap()

	if allBurnedLen != totalBookCount {
		t.Errorf(
			"Should have burned %d, but got %d. %d more to burn.",
			totalBookCount,
			allBurnedLen,
			totalBookCount-allBurnedLen,
		)
	}

	if burnedIDMapLen != totalBookCount {
		t.Errorf("Should have burned %d, but got %d", totalBookCount, burnedIDMapLen)
	}

	if incineratorMapLen != incineratorCount {
		t.Errorf(
			"Should have %d incinerators, but got %d",
			incineratorCount,
			incineratorMapLen,
		)
	}

	if providerMapLen != gopherCount {
		t.Errorf("Should have %d gophers, but got %d", gopherCount, providerMapLen)
	}

	for key := range burnedIDMap {
		taken := takenMap[key]
		provided := providerMap[key]

		if taken != provided {
			t.Errorf("Taken %d is different from provided %d", taken, provided)
		}
	}
}
