package goburnbooks

import (
	"testing"
	"time"
)

type testParams struct {
	burnDuration time.Duration
}

// This function is specific only to incinerator tests. Please do not use it
// anywhere else.
func totalBurnCountForAllRounds(ts *TestSuite) uint {
	return ts.burnRounds * ts.gopherCount * ts.supplyPerPileCount
}

func Test_BurnMultiple_ShouldEventuallyBurnAll(t *testing.T) {
	/// Setup
	t.Parallel()
	suite := NewDefaultTestSuite()
	incinerators := suite.Incinerators()
	providers := suite.BurnableProviders()
	totalBurnCount := int(totalBurnCountForAllRounds(suite))

	igParams := IncineratorGroupParams{
		BurnResultCapacity: uint(totalBurnCount),
		Incinerators:       incinerators,
	}

	ig := NewIncineratorGroup(&igParams)

	/// When
	for _, provider := range providers {
		ig.Consume(provider)
	}

	time.Sleep(suite.waitDuration)

	// Then
	contribPercentThreshold := suite.contribPercentThreshold
	verifyIncGroupFairContrib(ig, contribPercentThreshold, t)
	allBurned := ig.Burned()
	allBurnedLen := len(allBurned)
	burnedMap := ig.BurnedIDMap()
	burnedMapLen := len(burnedMap)

	if allBurnedLen != totalBurnCount {
		t.Errorf("Should have burned %d, but got %d", totalBurnCount, allBurnedLen)
	}

	for key, value := range burnedMap {
		if value != 1 {
			t.Errorf("%s should have been burned once, but got %d", key, value)
		}
	}

	if burnedMapLen != totalBurnCount {
		t.Errorf("Should have burned %d, but got %d", totalBurnCount, allBurnedLen)
	}
}

func Test_BurnMultiple_ShouldCapAtSpecifiedCapacity(t *testing.T) {
	/// Setup
	t.Parallel()

	suite := NewDefaultTestSuite()

	// Unrealistic burn duration to simulate blocking process.
	suite.burnDuration = 1e15
	suite.incineratorCount = 1
	incinerators := suite.Incinerators()
	providers := suite.BurnableProviders()
	totalBurnCount := int(totalBurnCountForAllRounds(suite))

	igParams := IncineratorGroupParams{
		BurnResultCapacity: uint(totalBurnCount),
		Incinerators:       incinerators,
	}

	ig := NewIncineratorGroup(&igParams)

	/// When
	for _, provider := range providers {
		ig.Consume(provider)
	}

	time.Sleep(suite.waitDuration)

	/// Then
	contribPercentThreshold := suite.contribPercentThreshold
	verifyIncGroupFairContrib(ig, contribPercentThreshold, t)
	burnedLength := len(ig.Burned())

	if burnedLength != 0 {
		t.Errorf("Should not have burned anything, but got %d", burnedLength)
	}
}
