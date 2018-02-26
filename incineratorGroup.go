package goburnbooks

// IncineratorGroup represents a group of incinerators.
type IncineratorGroup interface {
	Incinerator
	Burned() []*BurnResult
}

type incineratorGroup struct {
	availableIncinerators chan chan<- Burnable
	burned                []*BurnResult
	burnResult            chan *BurnResult
	incinerators          []FIncinerator
	updateBurned          chan *BurnResult
}

func (ig *incineratorGroup) Burned() []*BurnResult {
	return ig.burned
}

// We need to check if any incinerator is ready first before we can stack up
// Burnables for burning.
func (ig *incineratorGroup) Incinerate(burnables ...Burnable) {
	go func() {
		select {
		case ch := <-ig.availableIncinerators:
			for _, burnable := range burnables {
				ch <- burnable
			}
		}
	}()
}

// Loop each incinerator to fetch burned updates.
func (ig *incineratorGroup) loopBurn() {
	for _, i := range ig.incinerators {
		go func(i FIncinerator) {
			for {
				select {
				case burned := <-i.BurnResult():
					ig.updateBurned <- burned

				case <-i.SignalAvailable():
					go func() {
						ig.availableIncinerators <- i.Pending()
					}()
				}
			}
		}(i)
	}

	go func() {
		for {
			select {
			case burned := <-ig.updateBurned:
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
		availableIncinerators: make(chan chan<- Burnable, len(incinerators)),
		burned:                make([]*BurnResult, 0),
		burnResult:            make(chan *BurnResult),
		incinerators:          incinerators,
		updateBurned:          make(chan *BurnResult),
	}

	ig.loopBurn()
	return ig
}
