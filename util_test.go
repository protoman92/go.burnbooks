package goburnbooks

import (
	"math"
	"testing"
)

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
	spg SupplyPileGroup,
	percentThreshold float64,
	t *testing.T,
) {
	spContribMap := spg.SupplyPileContribMap()
	takerContribMap := spg.SupplyTakerContribMap()
	verifyFairContribution(spContribMap, percentThreshold, t)
	verifyFairContribution(takerContribMap, percentThreshold, t)
}

func verifyIncGroupFairContrib(
	ig IncineratorGroup,
	percentThreshold float64,
	t *testing.T,
) {
	incContribMap := ig.IncineratorContribMap()
	providerContribMap := ig.ProviderContribMap()
	verifyFairContribution(incContribMap, percentThreshold, t)
	verifyFairContribution(providerContribMap, percentThreshold, t)
}
