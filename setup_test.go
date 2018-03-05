package goburnbooks

import (
	"fmt"
	"strconv"
	"time"
)

type TestPlayers struct {
	books            []Book
	bookIds          []string
	gophers          []Gopher
	incinerators     []FIncinerator
	incineratorGroup IncineratorGroup
	supplyPiles      []FSupplyPile
	supplyPileGroup  SupplyPileGroup
}

func (tp *TestPlayers) GopherCount() int {
	return len(tp.gophers)
}

func (tp *TestPlayers) IncineratorCount() int {
	return len(tp.incinerators)
}

func (tp *TestPlayers) BookCount() int {
	return len(tp.bookIds)
}

type TestSuite struct {
	burnDuration            time.Duration
	burnRounds              uint
	contribPercentThreshold float64
	gopherCapacity          uint
	gopherCount             uint
	gopherTakeTimeout       time.Duration
	incineratorCap          uint
	incineratorCount        uint
	incineratorMinCap       uint
	integrationWaitDuration time.Duration
	logger                  Logger
	supplyPerPileCount      uint
	supplyPileCount         uint
	supplyPileTimeout       time.Duration
	tripDelay               time.Duration
	waitDuration            time.Duration
}

func (ts *TestSuite) TotalSupplyCount() uint {
	return ts.supplyPerPileCount * ts.supplyPileCount
}

func (ts *TestSuite) Gophers() []Gopher {
	gophers := make([]Gopher, ts.gopherCount)

	for ix := range gophers {
		gParams := GopherParams{
			BurnableProviderRawParams: BurnableProviderRawParams{
				BPID: strconv.Itoa(ix),
			},
			SupplyTakerRawParams: SupplyTakerRawParams{
				Cap:         ts.gopherCapacity,
				STID:        strconv.Itoa(ix),
				TakeTimeout: ts.gopherTakeTimeout,
			},
			Logger:       ts.logger,
			TripDuration: ts.tripDelay,
		}

		gopher := NewGopher(&gParams)
		gophers[ix] = gopher
	}

	return gophers
}

func (ts *TestSuite) SupplyPiles() ([]FSupplyPile, []Book, []string) {
	piles := make([]FSupplyPile, ts.supplyPileCount)
	allBooks := make([]Book, 0)
	allBookIds := make([]string, 0)

	for ix := range piles {
		supplies := make([]Suppliable, ts.supplyPerPileCount)

		for jx := range supplies {
			id := fmt.Sprintf("%d-%d", ix, jx)
			bParams := BookParams{BurnDuration: ts.burnDuration, ID: id}
			book := NewBook(&bParams)
			supplies[jx] = book
			allBooks = append(allBooks, book)
			allBookIds = append(allBookIds, id)
		}

		pParams := SupplyPileParams{
			Logger:             ts.logger,
			Supply:             supplies,
			ID:                 strconv.Itoa(ix),
			TakeResultCapacity: 0,
			TakeTimeout:        ts.supplyPileTimeout,
		}

		pile := NewSupplyPile(&pParams)
		piles[ix] = pile
	}

	return piles, allBooks, allBookIds
}

func (ts *TestSuite) Incinerators() []FIncinerator {
	incinerators := make([]FIncinerator, ts.incineratorCount)

	for ix := range incinerators {
		iParams := IncineratorParams{
			Capacity:    ts.incineratorCap,
			ID:          strconv.Itoa(ix),
			Logger:      ts.logger,
			MinCapacity: ts.incineratorMinCap,
		}

		incinerator := NewIncinerator(&iParams)
		incinerators[ix] = incinerator
	}

	return incinerators
}

func (ts *TestSuite) BurnableProviders() []BurnableProvider {
	providers := make([]BurnableProvider, ts.gopherCount)

	for pix := range providers {
		provideCh := make(chan []Burnable)
		prRawParams := BurnableProviderRawParams{BPID: strconv.Itoa(pix)}
		readyCh := make(chan interface{})

		prParams := BurnableProviderParams{
			BurnableProviderRawParams: prRawParams,
			BPLogger:                  ts.logger,
			ReceiveBurnableSourceCh:   provideCh,
		}

		provider := NewBurnableProvider(&prParams)

		go func(ix int) {
			for j := 0; j < int(ts.burnRounds); j++ {
				burnables := make([]Burnable, ts.supplyPerPileCount)

				for bix := range burnables {
					bParams := BookParams{
						BurnDuration: ts.burnDuration,
						ID:           fmt.Sprintf("%d-%d-%d", ix, j, bix),
					}

					burnable := NewBook(&bParams)
					burnables[bix] = burnable
				}

				provideCh <- burnables
				time.Sleep(1e5)
			}
		}(pix)

		go func() {
			for {
				<-readyCh
			}
		}()

		providers[pix] = provider
	}

	return providers
}

func (ts *TestSuite) SupplyTakers() []SupplyTaker {
	supplyTakers := make([]SupplyTaker, ts.gopherCount)
	totalSupplyCount := ts.TotalSupplyCount()

	for ix := range supplyTakers {
		readyCh := make(chan interface{})

		stRawParams := SupplyTakerRawParams{
			Cap:  ts.gopherCapacity,
			STID: strconv.Itoa(ix),
		}

		stParams := &SupplyTakerParams{
			SupplyTakerRawParams: stRawParams,
			SendSupplyDestCh:     make(chan []Suppliable, totalSupplyCount),
			STLogger:             ts.logger,
		}

		// Assume that the supply taker takes repeatedly at a specified delay.
		go func() {
			for {
				time.Sleep(1e5)
				readyCh <- true
			}
		}()

		supply := NewSupplyTaker(stParams)
		supplyTakers[ix] = supply
	}

	return supplyTakers
}

func (ts *TestSuite) SetUpSystem() *TestPlayers {
	gophers := ts.Gophers()
	piles, books, bookIds := ts.SupplyPiles()
	pileGroup := NewSupplyPileGroup(piles...)
	incinerators := ts.Incinerators()
	totalSupplyCount := ts.TotalSupplyCount()

	igParams := IncineratorGroupParams{
		BurnResultCapacity: totalSupplyCount,
		Incinerators:       incinerators,
	}

	incineratorGroup := NewIncineratorGroup(&igParams)

	for _, gopher := range gophers {
		go pileGroup.Supply(gopher)
		go incineratorGroup.Consume(gopher)
	}

	return &TestPlayers{
		books:            books,
		bookIds:          bookIds,
		gophers:          gophers,
		incinerators:     incinerators,
		incineratorGroup: incineratorGroup,
		supplyPiles:      piles,
		supplyPileGroup:  pileGroup,
	}
}

func NewDefaultTestSuite() *TestSuite {
	incineratorCap := uint(8)

	return &TestSuite{
		burnDuration:            time.Duration(1e5),
		burnRounds:              10,
		contribPercentThreshold: 0.2,
		gopherCapacity:          19,
		gopherCount:             5,
		gopherTakeTimeout:       time.Duration(1e5),
		incineratorCap:          incineratorCap,
		incineratorCount:        5,
		incineratorMinCap:       incineratorCap / 2,
		integrationWaitDuration: time.Duration(8e9),
		logger:                  NewLogger(false),
		supplyPerPileCount:      1000,
		supplyPileCount:         5,
		supplyPileTimeout:       time.Duration(1e5),
		tripDelay:               time.Duration(1e8),
		waitDuration:            time.Duration(5e9),
	}
}
