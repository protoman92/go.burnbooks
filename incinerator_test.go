package goburnbooks

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

type testParams struct {
	burnDuration time.Duration
}

func newBurnables(params testParams, pref1 string, pref2 string) []Burnable {
	burnables := make([]Burnable, burnableCountPerRound)

	for ix := range burnables {
		bParams := BookParams{
			BurnDuration: params.burnDuration,
			ID:           fmt.Sprintf("%s-%s-%d", pref1, pref2, ix),
		}

		burnable := NewBook(&bParams)
		burnables[ix] = burnable
	}

	return burnables
}

func incinerate(ig IncineratorGroup, params testParams) {
	for i := 0; i < providerCount; i++ {
		provideCh := make(chan []Burnable)
		prRawParams := BurnableProviderRawParams{BPID: strconv.Itoa(i)}
		readyCh := make(chan interface{})

		prParams := BurnableProviderParams{
			BurnableProviderRawParams: prRawParams,
			ProvideCh:                 provideCh,
			ConsumeReadyCh:            readyCh,
		}

		provider := NewBurnableProvider(&prParams)

		go func(ix int) {
			for j := 0; j < burnRounds; j++ {
				burnables := newBurnables(params, strconv.Itoa(ix), strconv.Itoa(j))
				provideCh <- burnables
				time.Sleep(1e5)
			}
		}(i)

		go func() {
			for {
				<-readyCh
			}
		}()

		ig.Consume(provider)
	}

	time.Sleep(waitDuration)
}

func Test_BurnMultiple_ShouldEventuallyBurnAll(t *testing.T) {
	/// Setup
	t.Parallel()

	iParams := &IncineratorParams{
		Capacity:    incineratorCap,
		ID:          "1",
		Logger:      logWorker,
		MinCapacity: incineratorMinCap,
	}

	incinerator := NewIncinerator(iParams)
	ig := NewIncineratorGroup(incinerator)

	/// When
	incinerate(ig, testParams{burnDuration: 1e5})

	// Then
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

	iParams := &IncineratorParams{
		Capacity:    incineratorCap,
		ID:          "1",
		Logger:      logWorker,
		MinCapacity: incineratorMinCap,
	}

	incinerator := NewIncinerator(iParams)
	ig := NewIncineratorGroup(incinerator)

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
	allIncs := make([]Incinerator, incineratorCount)

	for ix := range allIncs {
		id := strconv.Itoa(ix)

		iParams := &IncineratorParams{
			Capacity:    incineratorCap,
			ID:          id,
			Logger:      logWorker,
			MinCapacity: incineratorMinCap,
		}

		allIncs[ix] = NewIncinerator(iParams)
	}

	ig := NewIncineratorGroup(allIncs...)

	/// When
	incinerate(ig, testParams{burnDuration: 1e5})

	/// Then
	verifyIncGroupFairContrib(ig, contribPercentThreshold, t)
	burnIDMap := ig.BurnedIDMap()
	burnedIDMapLen := len(burnIDMap)
	incineratorMap := ig.IncineratorContribMap()
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

	if burnedIDMapLen != totalBurnCount {
		t.Errorf("Should have burned %d, but got %d", totalBurnCount, burnedIDMapLen)
	}
}
