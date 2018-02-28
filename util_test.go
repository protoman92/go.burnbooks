package goburnbooks_test

import (
	gbb "goburnbooks"
	"math"
	"testing"
)

func burnedIDMap(ig gbb.IncineratorGroup) map[string]int {
	allBurned := ig.Burned()
	burnedMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.Burned.BurnableID()
		burnedMap[id] = burnedMap[id] + 1
	}

	return burnedMap
}

func supplyPileTakenContribMap(pg gbb.SupplyPileGroup) map[string]int {
	allTaken := pg.Taken()
	contributorMap := make(map[string]int, 0)

	for _, taken := range allTaken {
		id := taken.PileID
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

func supplyTakerTakenContribMap(pg gbb.SupplyPileGroup) map[string]int {
	allTaken := pg.Taken()
	contributorMap := make(map[string]int, 0)

	for _, taken := range allTaken {
		id := taken.TakerID
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

func incineratorBurnedContribMap(ig gbb.IncineratorGroup) map[string]int {
	allBurned := ig.Burned()
	contributorMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.IncineratorID
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

func providerBurnedContribMap(ig gbb.IncineratorGroup) map[string]int {
	allBurned := ig.Burned()
	contributorMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.ProviderID
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

// If the system is written correctly, contributions should not deviate too
// much from each other.
func verifyFairContribution(
	contrib map[string]int,
	percentThreshold float64,
	t *testing.T,
) {
	contribLen := len(contrib)

	if contribLen == 0 {
		return
	}

	totalContrib := 0

	for _, value := range contrib {
		totalContrib += value
	}

	average := float64(totalContrib / contribLen)

	for key, value := range contrib {
		diff := math.Abs(float64(value) - average)

		if average > 0 {
			percentDiff := float64(diff / average)

			if percentDiff > percentThreshold {
				t.Errorf(
					`%s's contribution difference exceeded threshold %.2f%% (was %.2f%%).
					Actual contribution was %d, while average was %.2f`,
					key,
					percentThreshold*100,
					percentDiff*100,
					value,
					average,
				)
			}
		}
	}
}

func verifySupplyGroupFairContrib(
	spg gbb.SupplyPileGroup,
	percentThreshold float64,
	t *testing.T,
) {
	spContribMap := supplyPileTakenContribMap(spg)
	takerContribMap := supplyTakerTakenContribMap(spg)
	verifyFairContribution(spContribMap, percentThreshold, t)
	verifyFairContribution(takerContribMap, percentThreshold, t)
}

func verifyIncGroupFairContrib(
	ig gbb.IncineratorGroup,
	percentThreshold float64,
	t *testing.T,
) {
	incContribMap := incineratorBurnedContribMap(ig)
	providerContribMap := providerBurnedContribMap(ig)
	verifyFairContribution(incContribMap, percentThreshold, t)
	verifyFairContribution(providerContribMap, percentThreshold, t)
}
