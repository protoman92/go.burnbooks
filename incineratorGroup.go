package goburnbooks

// IncineratorGroup represents a group of incinerators.
type IncineratorGroup interface {
	Incinerator
	Burned() []*BurnResult

	// Get the id's of all burned Burnables.
	BurnedIDMap() map[string]int

	// Get the contributions (i.e. burn count) of each incinerator.
	IncineratorContribMap() map[string]int

	// Get the contributions (i.e. burnable provision count) of each provider.
	ProviderContribMap() map[string]int
}

type incineratorGroup struct {
	burned       []*BurnResult
	burnResultCh chan *BurnResult
	incinerators []Incinerator
}

func (ig *incineratorGroup) Burned() []*BurnResult {
	return ig.burned
}

func (ig *incineratorGroup) BurnedIDMap() map[string]int {
	allBurned := ig.burned
	burnedMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.Burned.BurnableID()
		burnedMap[id] = burnedMap[id] + 1
	}

	return burnedMap
}

func (ig *incineratorGroup) IncineratorContribMap() map[string]int {
	allBurned := ig.burned
	contributorMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.IncineratorID
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

func (ig *incineratorGroup) ProviderContribMap() map[string]int {
	allBurned := ig.burned
	contributorMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.ProviderID
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}

func (ig *incineratorGroup) BurnResultChannel() <-chan *BurnResult {
	return ig.burnResultCh
}

func (ig *incineratorGroup) Consume(provider BurnableProvider) {
	for _, i := range ig.incinerators {
		go i.Consume(provider)
	}
}

func (ig *incineratorGroup) UID() string {
	var id string

	for _, i := range ig.incinerators {
		id += id + "-" + i.UID()
	}

	return id
}

// Loop each incinerator to fetch burned updates.
func (ig *incineratorGroup) loopBurn() {
	updateAllBurnedCh := make(chan *BurnResult)

	for _, i := range ig.incinerators {
		go func(i Incinerator) {
			var burnResultCh = i.BurnResultChannel()
			var burnResult *BurnResult
			var updateBurnedCh chan<- *BurnResult

			for {
				// The sequence of this statement is:
				// - When we receive a new burn result, set the burn result channel to
				// nil to process in peace. Afterwards, initialize the burn result and
				// result update channel.
				// - After the burn result has been updated, reinstate the burn result
				// channel to keep receiving updates.
				select {
				case burned := <-burnResultCh:
					burnResultCh = nil
					burnResult = burned
					updateBurnedCh = updateAllBurnedCh

				case updateBurnedCh <- burnResult:
					burnResult = nil
					updateBurnedCh = nil
					burnResultCh = i.BurnResultChannel()
				}
			}
		}(i)
	}

	go func() {
		for {
			select {
			case burned := <-updateAllBurnedCh:
				ig.burned = append(ig.burned, burned)

				go func() {
					ig.burnResultCh <- burned
				}()
			}
		}
	}()
}

// NewIncineratorGroup creates a new incinerator group from a number of
// incinerators. An incinerator group implements the same functionalities as
// an incinerator, so we can access them directly instead of viewing individual
// incinerators.
func NewIncineratorGroup(incinerators ...Incinerator) IncineratorGroup {
	ig := &incineratorGroup{
		burned:       make([]*BurnResult, 0),
		burnResultCh: make(chan *BurnResult),
		incinerators: incinerators,
	}

	go ig.loopBurn()
	return ig
}
