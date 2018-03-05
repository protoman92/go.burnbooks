package goburnbooks

import (
	"sync"
)

// IncineratorGroup represents a group of incinerators.
type IncineratorGroup interface {
	Incinerator
	Burned() []BurnResult

	// Get the id's of all burned Burnables.
	BurnedIDMap() map[string]int

	// Get the contributions (i.e. burn count) of each incinerator.
	IncineratorContribMap() map[string]int

	// Get the contributions (i.e. burnable provision count) of each provider.
	ProviderContribMap() map[string]int
}

// IncineratorGroupParams represents all the required parameters to build an
// IncineratorGroup.
type IncineratorGroupParams struct {
	Incinerators       []FIncinerator
	BurnResultCapacity uint
}

type incineratorGroup struct {
	IncineratorGroupParams
	mutex        sync.RWMutex
	burned       []BurnResult
	burnResultCh chan BurnResult
}

func (ig *incineratorGroup) Burned() []BurnResult {
	ig.mutex.RLock()
	defer ig.mutex.RUnlock()
	return ig.burned
}

func (ig *incineratorGroup) BurnedIDMap() map[string]int {
	ig.mutex.RLock()
	allBurned := ig.burned
	ig.mutex.RUnlock()

	burnedMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.Burned().BurnableID()
		burnedMap[id] = burnedMap[id] + 1
	}

	return burnedMap
}

func (ig *incineratorGroup) IncineratorContribMap() map[string]int {
	ig.mutex.RLock()
	allBurned := ig.burned
	ig.mutex.RUnlock()

	contributorMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.IncineratorID()
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

func (ig *incineratorGroup) ProviderContribMap() map[string]int {
	ig.mutex.RLock()
	allBurned := ig.burned
	ig.mutex.RUnlock()

	contributorMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.ProviderID()
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

func (ig *incineratorGroup) BurnResultChannel() <-chan BurnResult {
	return ig.burnResultCh
}

func (ig *incineratorGroup) Consume(provider BurnableProvider) {
	for _, i := range ig.Incinerators {
		go i.Consume(provider)
	}
}

func (ig *incineratorGroup) UID() string {
	var id string

	for _, i := range ig.Incinerators {
		id += id + "-" + i.UID()
	}

	return id
}

// Loop each incinerator to fetch burned updates.
func (ig *incineratorGroup) loopBurn() {
	updateAllBurnedCh := make(chan BurnResult)

	for _, i := range ig.Incinerators {
		go func(i FIncinerator) {
			resetSequenceCh := make(chan interface{}, 1)
			var burnResultCh = i.BurnResultChannel()
			var burnResult BurnResult
			var updateBurnedCh chan<- BurnResult

			for {
				// The sequence of this statement is:
				// - When we receive a new burn result, set the burn result channel to
				// nil to process in peace. Afterwards, initialize the burn result and
				// result update channel.
				// - After the burn result has been updated, reinstate the burn result
				// channel to keep receiving updates.
				select {
				case burned, ok := <-burnResultCh:
					burnResultCh = nil

					if ok {
						burnResult = burned
						updateBurnedCh = updateAllBurnedCh
					} else {
						return
					}

				case updateBurnedCh <- burnResult:
					burnResult = nil
					updateBurnedCh = nil
					resetSequenceCh <- true

				case <-resetSequenceCh:
					burnResultCh = i.BurnResultChannel()
				}
			}
		}(i)
	}

	go func() {
		updateBurned := updateAllBurnedCh
		var burnResultCh chan<- BurnResult
		var lastBurned BurnResult

		for {
			select {
			case burned := <-updateBurned:
				updateBurned = nil

				// Note that this mutex is only used to modify the burned result
				// map, since said map is accessible via a getter method.
				ig.mutex.Lock()
				ig.burned = append(ig.burned, burned)
				ig.mutex.Unlock()
				lastBurned = burned
				burnResultCh = ig.burnResultCh

			case burnResultCh <- lastBurned:
				burnResultCh = nil
				lastBurned = nil
				updateBurned = updateAllBurnedCh
			}
		}
	}()
}

// NewIncineratorGroup creates a new incinerator group from a number of
// incinerators. An incinerator group implements the same functionalities as
// an incinerator, so we can access them directly instead of viewing individual
// incinerators.
func NewIncineratorGroup(params *IncineratorGroupParams) IncineratorGroup {
	ig := &incineratorGroup{
		IncineratorGroupParams: *params,
		burned:                 make([]BurnResult, 0),
		burnResultCh:           make(chan BurnResult, params.BurnResultCapacity),
	}

	go ig.loopBurn()
	return ig
}
