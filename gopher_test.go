package goburnbooks

import (
	"testing"
	"time"
)

func Test_GopherDeliveringBurnables_ShouldBurnAll(t *testing.T) {
	/// Setup
	t.Parallel()
	suite := NewDefaultTestSuite()

	/// When
	players := suite.SetUpSystem()
	time.Sleep(suite.integrationWaitDuration)

	/// Then
	pileGroup := players.supplyPileGroup
	contribPercentThreshold := suite.contribPercentThreshold
	gopherCount := int(players.GopherCount())
	incineratorCount := int(players.IncineratorCount())
	incineratorGroup := players.incineratorGroup
	totalBookCount := int(players.BookCount())
	totalSupplyCount := int(suite.TotalSupplyCount())
	verifySupplyGroupFairContrib(pileGroup, contribPercentThreshold, t)
	verifyIncGroupFairContrib(incineratorGroup, contribPercentThreshold, t)

	if totalBookCount != totalSupplyCount {
		t.Errorf("Should have %d books, but got %d", totalSupplyCount, totalBookCount)
	}

	allBurned := incineratorGroup.Burned()
	allBurnedLen := len(allBurned)
	burnedIDMap := incineratorGroup.BurnedIDMap()
	burnedIDMapLen := len(burnedIDMap)
	incineratorMap := incineratorGroup.IncineratorContribMap()
	incineratorMapLen := len(incineratorMap)
	incineratorBurned := totalContribCount(incineratorMap)
	providerMap := incineratorGroup.ProviderContribMap()
	providerMapLen := len(providerMap)
	supplyPileContribMap := pileGroup.SupplyPileContribMap()
	supplyProvided := totalContribCount(supplyPileContribMap)
	takenMap := pileGroup.SupplyTakerContribMap()

	if allBurnedLen != totalBookCount {
		t.Errorf(
			"Should have burned %d, but got %d. %d more to burn.",
			totalBookCount,
			allBurnedLen,
			totalBookCount-allBurnedLen,
		)
	}

	if supplyProvided != incineratorBurned {
		t.Errorf("Supplied %d and burned %d", supplyProvided, incineratorBurned)
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
