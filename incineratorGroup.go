package goburnbooks

// IncineratorGroup represents a group of incinerators.
type IncineratorGroup interface {
	Incinerator
	Burned() []*BurnResult
}

type incineratorGroup struct {
	availableIncinerators chan chan<- []Burnable
	burned                []*BurnResult
	incinerators          []FIncinerator
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
			ch <- burnables
		}
	}()
}

// Loop each incinerator to fetch burned updates.
func (ig *incineratorGroup) loopBurn() {
	updateBurned := make(chan *BurnResult)

	for _, i := range ig.incinerators {
		go func(i FIncinerator) {
			availability := i.Availability()
			burnResult := i.BurnResult()

			for {
				select {
				case burned := <-burnResult:
					go func() {
						updateBurned <- burned
					}()

				case ch := <-availability:
					go func() {
						ig.availableIncinerators <- ch
					}()
				}
			}
		}(i)
	}

	go func() {
		for {
			select {
			case burned := <-updateBurned:
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
		availableIncinerators: make(chan chan<- []Burnable, len(incinerators)),
		burned:                make([]*BurnResult, 0),
		incinerators:          incinerators,
	}

	ig.loopBurn()
	return ig
}
