package goburnbooks_test

import (
	"time"

	gbb "github.com/protoman92/goburnbooks"
)

const (
	burnableCountPerRound = 1000
	burnRounds            = 30
	providerCount         = 5
	totalBurnCount        = providerCount * burnableCountPerRound * burnRounds
)

const (
	supplyTakerCount = 10
)

const (
	contribPercentThreshold = 0.2
	gopherCapacity          = 30
	gopherCount             = 3
	gopherTakeTimeout       = time.Duration(1e5)
	incineratorCap          = 30
	incineratorCount        = 3
	incineratorMinCap       = incineratorCap / 2
	integrationWaitDuration = time.Duration(10e9)
	supplyPerPileCount      = 1000
	supplyPileCount         = 5
	supplyPileTimeout       = time.Duration(1e5)
	tripDelay               = time.Duration(1e8)
	totalSupplyCount        = supplyPileCount * supplyPerPileCount
	waitDuration            = time.Duration(5e9)
)

var (
	logger = gbb.NewLogger(false)
)
