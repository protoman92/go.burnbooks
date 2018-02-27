package goburnbooks_test

const (
	burnableCountPerRound  = 1000
	burnRounds             = 30
	incineratorCap         = 100
	minIncineratorCapacity = incineratorCap / 2
	providerCount          = 5
	supplyPileTimeout      = 1e8
	totalBurnCount         = providerCount * burnableCountPerRound * burnRounds
	waitDuration           = 2e9
)
