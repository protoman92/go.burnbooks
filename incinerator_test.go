package goburnbooks_test

import (
	"fmt"
	gbb "goburnbooks"
	"strconv"
	"testing"
	"time"
)

type testParams struct {
	burnDuration time.Duration
}

func newBurnables(params testParams, pref1 string, pref2 string) []gbb.Burnable {
	burnables := make([]gbb.Burnable, burnableCountPerRound)

	for ix := range burnables {
		bParams := &gbb.BookParams{
			BurnDuration: params.burnDuration,
			ID:           fmt.Sprintf("%s-%s-%d", pref1, pref2, ix),
		}

		burnable := gbb.NewBook(bParams)
		burnables[ix] = burnable
	}

	return burnables
}

func incinerate(ig gbb.IncineratorGroup, params testParams) {
	for i := 0; i < providerCount; i++ {
		provideCh := make(chan []gbb.Burnable)
		prRawParams := &gbb.BurnableProviderRawParams{BPID: strconv.Itoa(i)}

		prParams := &gbb.BurnableProviderParams{
			BurnableProviderRawParams: prRawParams,
			ProvideCh:                 provideCh,
			ConsumeReadyCh:            make(chan interface{}, 1),
		}

		provider := gbb.NewBurnableProvider(prParams)

		go func(ix int) {
			for j := 0; j < burnRounds; j++ {
				burnables := newBurnables(params, strconv.Itoa(ix), strconv.Itoa(j))
				provideCh <- burnables
				time.Sleep(1e5)
			}
		}(i)

		ig.Consume(provider)
	}

	time.Sleep(waitDuration)
}

func Test_BurnMultiple_ShouldEventuallyBurnAll(t *testing.T) {
	/// Setup
	t.Parallel()

	iParams := &gbb.IncineratorParams{
		Capacity:    incineratorCap,
		ID:          "1",
		Logger:      logger,
		MinCapacity: incineratorMinCap,
	}

	incinerator := gbb.NewIncinerator(iParams)
	ig := gbb.NewIncineratorGroup(incinerator)

	/// When
	incinerate(ig, testParams{burnDuration: 1e5})

	// Then
	verifyIncGroupFairContrib(ig, contribPercentThreshold, t)
	allBurned := ig.Burned()
	allBurnedLen := len(allBurned)
	burnedMap := burnedIDMap(ig)
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

	iParams := &gbb.IncineratorParams{
		Capacity:    incineratorCap,
		ID:          "1",
		Logger:      logger,
		MinCapacity: incineratorMinCap,
	}

	incinerator := gbb.NewIncinerator(iParams)
	ig := gbb.NewIncineratorGroup(incinerator)

	/// When
	// Unrealistic burn duration to represent blocking operation.
	incinerate(ig, testParams{burnDuration: 1e15})

	/// Then
	verifyIncGroupFairContrib(ig, contribPercentThreshold, t)
	burnedLength := len(ig.Burned())

	if burnedLength != 0 {
		t.Errorf("Should not have burned anything, but got %d", burnedLength)
	}
}

func Test_BurnMultipleBooksWithIncineratorGroup_ShouldAllocate(t *testing.T) {
	/// Setup
	t.Parallel()
	allIncs := make([]gbb.FIncinerator, incineratorCount)

	for ix := range allIncs {
		id := strconv.Itoa(ix)

		iParams := &gbb.IncineratorParams{
			Capacity:    incineratorCap,
			ID:          id,
			Logger:      logger,
			MinCapacity: incineratorMinCap,
		}

		allIncs[ix] = gbb.NewIncinerator(iParams)
	}

	ig := gbb.NewIncineratorGroup(allIncs...)

	/// When
	incinerate(ig, testParams{burnDuration: 1e5})

	/// Then
	verifyIncGroupFairContrib(ig, contribPercentThreshold, t)
	burnIdMap := burnedIDMap(ig)
	burnedIdMapLen := len(burnIdMap)
	incineratorMap := incineratorBurnedContribMap(ig)
	incineratorMapLen := len(incineratorMap)

	if incineratorMapLen != incineratorCount {
		t.Errorf(
			"Expected %d incinerators, but got %d",
			incineratorCount,
			incineratorMapLen,
		)
	}

	for key, value := range incineratorMap {
		if value == 0 {
			t.Errorf("%s should have burned something, but got nothing", key)
		}
	}

	if burnedIdMapLen != totalBurnCount {
		t.Errorf("Should have burned %d, but got %d", totalBurnCount, burnedIdMapLen)
	}
}
