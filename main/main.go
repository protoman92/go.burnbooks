package main

import (
	"fmt"
	"strconv"
	"time"

	gbb "goburnbooks"
)

const (
	burnDuration       = time.Duration(1e8)
	gopherCount        = 10
	gopherCapacity     = 23
	gopherTakeTimeout  = time.Duration(1e5)
	incineratorCap     = 20
	incineratorMinCap  = incineratorCap / 2
	incineratorCount   = 6
	tripDelay          = time.Duration(1e9)
	supplyPerPileCount = 10000
	supplyPileCount    = 5
	supplyPileTimeout  = time.Duration(1e5)
	totalSupplyCount   = supplyPileCount * supplyPerPileCount
)

var (
	logger = gbb.NewLogger(false)
)

func main() {
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

	// Start the system
	for _, gopher := range gophers {
		go pileGroup.Supply(gopher)
		go incineratorGroup.Consume(gopher)
	}

	done := make(chan bool, 1)
	var totalBurnCount int

	go func() {
		burnResultCh := incineratorGroup.BurnResultChannel()
		var initLastBurned bool
		var lastBurned *gbb.BurnResult

		for {
			select {
			case result := <-burnResultCh:
				fmt.Printf("%v\n", result)

				if initLastBurned && result == lastBurned {
					panic(fmt.Sprintf("Should not have duplicate result %v", result))
				}

				lastBurned = result

				if !initLastBurned {
					initLastBurned = true
				}

				totalBurnCount++

				if totalBurnCount == totalSupplyCount {
					done <- true
				}
			}
		}
	}()

	select {
	case <-done:
		fmt.Printf("Burned a total of %d books", totalBurnCount)
		incContrib := incineratorGroup.IncineratorContribMap()
		providerContrib := incineratorGroup.ProviderContribMap()
		pileContrib := pileGroup.SupplyPileContribMap()
		takerContrib := pileGroup.SupplyTakerContribMap()

		fmt.Println()
		fmt.Printf("\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")

		for key, value := range pileContrib {
			fmt.Printf("Pile %s contributed %d books\n", key, value)
		}

		fmt.Printf("\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")

		for key, value := range incContrib {
			fmt.Printf("Incinerator %s burned %d book\n", key, value)
		}

		fmt.Printf("\n>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")

		for key, value := range providerContrib {
			taken := takerContrib[key]
			fmt.Printf("Gopher %s took %d and delivered %d books\n", key, taken, value)
		}
	}
}
