package goburnbooks

// IncineratorGroup represents a group of incinerators.
type IncineratorGroup interface {
	Incinerator
	Burned() []*BurnResult
}

type incineratorGroup struct {
	burned       []*BurnResult
	incinerators []FIncinerator
}

func (ig *incineratorGroup) Burned() []*BurnResult {
	return ig.burned
}

func (ig *incineratorGroup) Consume(provider BurnableProvider) {
	for _, i := range ig.incinerators {
		go i.Consume(provider)
	}
}

// Loop each incinerator to fetch burned updates.
func (ig *incineratorGroup) loopBurn() {
	updateAllBurnedCh := make(chan *BurnResult)

	for _, i := range ig.incinerators {
		go func(i FIncinerator) {
			var burnResultCh = i.BurnResult()
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
					burnResultCh = i.BurnResult()
				}
			}
		}(i)
	}

	go func() {
		for {
			select {
			case burned := <-updateAllBurnedCh:
				ig.burned = append(ig.burned, burned)
			}
		}
	}()
}

// NewIncineratorGroup creates a new incinerator group from a number of
// incinerators. An incinerator group implements the same functionalities as
// an incinerator, so we can access them directly instead of viewing individual
// incinerators.
func NewIncineratorGroup(incinerators ...FIncinerator) IncineratorGroup {
	ig := &incineratorGroup{
		burned:       make([]*BurnResult, 0),
		incinerators: incinerators,
	}

	go ig.loopBurn()
	return ig
}
