package goburnbooks_test

import (
	"fmt"
	gbb "goburnbooks"
	"strconv"
	"testing"
	"time"
)

func Test_GopherDeliveringBurnables_ShouldBurnAll(t *testing.T) {
	/// Setup
	t.Parallel()
	gophers := make([]gbb.Gopher, gopherCount)

	for ix := range gophers {
		gParams := &gbb.GopherParams{
			BurnableProviderRawParams: &gbb.BurnableProviderRawParams{
				BPID: strconv.Itoa(ix),
			},
			SupplyTakerRawParams: &gbb.SupplyTakerRawParams{
				Cap:  gopherCapacity,
				STID: strconv.Itoa(ix),
			},
			Logger:       logger,
			TakeTimeout:  gopherTakeTimeout,
			TripDuration: tripDelay,
		}

		gopher := gbb.NewGopher(gParams)
		gophers[ix] = gopher
	}

	burnDuration := time.Duration(1)
	piles := make([]gbb.SupplyPile, supplyPileCount)
	allBooks := make([]gbb.Book, 0)
	allBookIds := make([]string, 0)

	for ix := range piles {
		supplies := make([]gbb.Suppliable, supplyPerPileCount)

		for jx := range supplies {
			id := fmt.Sprintf("%d-%d", ix, jx)
			bParams := &gbb.BookParams{BurnDuration: burnDuration, ID: id}
			book := gbb.NewBook(bParams)
			supplies[jx] = book
			allBooks = append(allBooks, book)
			allBookIds = append(allBookIds, id)
		}

		pParams := &gbb.SupplyPileParams{
			Logger:      logger,
			Supply:      supplies,
			ID:          strconv.Itoa(ix),
			TakeTimeout: supplyPileTimeout,
		}

		pile := gbb.NewSupplyPile(pParams)
		piles[ix] = pile
	}

	pileGroup := gbb.NewSupplyPileGroup(piles...)

	incinerators := make([]gbb.Incinerator, incineratorCount)

	for ix := range incinerators {
		iParams := &gbb.IncineratorParams{
			Capacity:    incineratorCap,
			ID:          strconv.Itoa(ix),
			Logger:      logger,
			MinCapacity: incineratorMinCap,
		}

		incinerator := gbb.NewIncinerator(iParams)
		incinerators[ix] = incinerator
	}

	incineratorGroup := gbb.NewIncineratorGroup(incinerators...)
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
	allBookIdLen := len(allBookIds)

	if allBookIdLen != totalSupplyCount {
		t.Errorf("Should have %d books, but got %d", totalSupplyCount, allBookIdLen)
	}

	allBurned := incineratorGroup.Burned()
	allBurnedLen := len(allBurned)
	burnedIDMap := burnedIDMap(incineratorGroup)
	burnedIDMapLen := len(burnedIDMap)
	incineratorMap := incineratorBurnedContribMap(incineratorGroup)
	incineratorMapLen := len(incineratorMap)
	providerMap := providerBurnedContribMap(incineratorGroup)
	providerMapLen := len(providerMap)

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
}
