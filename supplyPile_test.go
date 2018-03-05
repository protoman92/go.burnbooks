package goburnbooks

import (
	"os"
	"testing"
	"time"
)

func Test_SupplyTakersHavingOddCapacity_ShouldStillLoadAll(t *testing.T) {
	/// Setup
	t.Parallel()
	suite := NewDefaultTestSuite()
	supplyPiles, _, _ := suite.SupplyPiles()
	totalSupplyCount := int(suite.TotalSupplyCount())
	pileGroup := NewSupplyPileGroup(supplyPiles...)
	supplyTakers := suite.SupplyTakers()

	/// When
	for _, taker := range supplyTakers {
		pileGroup.Supply(taker)
	}

	time.Sleep(suite.waitDuration)

	/// Then
	contribPercentThreshold := suite.contribPercentThreshold
	allTaken := pileGroup.Taken()
	verifySupplyGroupFairContrib(pileGroup, contribPercentThreshold, t)
	allTakenMap := make(map[string]int, 0)
	allTakenCount := 0

	for _, result := range allTaken {
		takerID := result.TakerID()
		takenCount := len(result.SupplyIDs())
		allTakenMap[takerID] = allTakenMap[takerID] + takenCount
		allTakenCount += takenCount
	}

	if allTakenCount != totalSupplyCount {
		t.Errorf("Should have taken %d, but got %d", totalSupplyCount, allTakenCount)
	}

	for key, value := range allTakenMap {
		if value == 0 {
			t.Errorf("%s should have taken some, but took nothing", key)
		}
	}

	os.Exit(0)
}
