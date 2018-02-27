package goburnbooks_test

import (
	"time"
)

const (
	burnableCountPerRound  = 1000
	burnRounds             = 1
	gopherCapacity         = 2
	gopherCount            = 2
	incineratorCap         = 100
	incineratorCount       = 1
	minIncineratorCapacity = incineratorCap / 2
	providerCount          = 1
	supplyPerPileCount     = 10
	supplyPileCount        = 1
	supplyPileTimeout      = time.Duration(1e8)
	tripDelay              = time.Duration(1e5)
	totalBurnCount         = providerCount * burnableCountPerRound * burnRounds
	totalSupplyCount       = supplyPileCount * supplyPerPileCount
	waitDuration           = time.Duration(2e9)
)

// const (
// 	burnableCountPerRound  = 1000
// 	burnRounds             = 30
// 	gopherCapacity         = 23
// 	gopherCount            = 10
// 	incineratorCap         = 100
// 	incineratorCount       = 10
// 	minIncineratorCapacity = incineratorCap / 2
// 	providerCount          = 5
// 	supplyPerPileCount     = 10000
// 	supplyPileCount        = 10
// 	supplyPileTimeout      = time.Duration(1e8)
// 	tripDelay              = time.Duration(1e5)
// 	totalBurnCount         = providerCount * burnableCountPerRound * burnRounds
// 	totalSupplyCount       = supplyPileCount * supplyPerPileCount
// 	waitDuration           = time.Duration(2e9)
// )
