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
			TripDuration: tripDelay,
		}

		gopher := gbb.NewGopher(gParams)
		gophers[ix] = gopher
	}

	burnDuration := time.Duration(1)
	piles := make([]gbb.FSupplyPile, supplyPileCount)
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

		pParams := &gbb.SupplyPileParams{Supply: supplies, ID: strconv.Itoa(ix)}
		pile := gbb.NewSupplyPile(pParams)
		piles[ix] = pile
	}

	pileGroup := gbb.NewSupplyPileGroup(piles...)

	incinerators := make([]gbb.FIncinerator, incineratorCount)

	for ix := range incinerators {
		iParams := &gbb.IncineratorParams{
			Capacity:    incineratorCap,
			ID:          strconv.Itoa(ix),
			MinCapacity: minIncineratorCapacity,
		}

		incinerator := gbb.NewIncinerator(iParams)
		incinerators[ix] = incinerator
	}

	incineratorGroup := gbb.NewIncineratorGroup(incinerators...)
	totalBookCount := len(allBookIds)
	t.Logf("Got %d piles of books", len(piles))
	t.Logf("Got %d gophers to deliver books", len(gophers))
	t.Logf("Got %d incinerators", len(incinerators))
	t.Logf("Got %d books to burn", totalBookCount)

	/// When
	for _, gopher := range gophers {
		go pileGroup.Supply(gopher)
		t.Logf("Supplying to gopher %v", gopher)
		go incineratorGroup.Consume(gopher)
		t.Logf("Consuming from gopher %v", gopher)
	}

	time.Sleep(waitDuration)

	/// Then
	allBookIdLen := len(allBookIds)

	if allBookIdLen != totalSupplyCount {
		t.Errorf("Should have %d books, but got %d", totalSupplyCount, allBookIdLen)
	}

	allBurned := incineratorGroup.Burned()
	allBurnedLen := len(allBurned)
	contributorMap := burnContributorMap(incineratorGroup)
	contributorMapLen := len(contributorMap)

	if allBurnedLen != totalBookCount {
		t.Errorf(
			"Should have burned %d, but got %d. %d more to burn.",
			totalBookCount,
			allBurnedLen,
			totalBookCount-allBurnedLen,
		)
	}

	for key, value := range contributorMap {
		t.Logf("Gopher %s burned %d books\n", key, value)
	}

	if contributorMapLen != gopherCount {
		t.Errorf(
			"Should have had %d gophers, but got %d",
			gopherCount,
			contributorMapLen,
		)
	}
}
