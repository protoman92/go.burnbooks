package goburnbooks

// IncineratorGroup represents a group of incinerators.
type IncineratorGroup interface {
	Incinerator
}

type incineratorGroup struct {
	availableIncinerators chan Incinerator
	burned                []Burner
	incinerators          []FullIncinerator
	updateBurned          chan []Burner
}

func (ig *incineratorGroup) Burned() []Burner {
	return ig.burned
}

// We need to check if any incinerator is ready first before we can stack up
// Burners for burning.
func (ig *incineratorGroup) Feed(burners ...Burner) {
	go func() {
		select {
		case i := <-ig.availableIncinerators:
			i.Feed(burners...)
		}
	}()
}

// Loop each incinerator to fetch burned updates.
func (ig *incineratorGroup) loopBurn() {
	for _, i := range ig.incinerators {
		go func(i FullIncinerator) {
			for {
				select {
				case available := <-i.Available():
					if available {
						go func() {
							ig.availableIncinerators <- i
						}()
					}
				}
			}
		}(i)

		go func(i FullIncinerator) {
			// Since this map is local to each incinerator, we do not need to worry
			// about concurrent modifications.
			burnedMap := make(map[Burner]bool)

			for {
				justBurned := make([]Burner, 0)

				for _, b := range i.Burned() {
					if !burnedMap[b] {
						justBurned = append(justBurned, b)
						burnedMap[b] = true
					}
				}

				go func() {
					ig.updateBurned <- justBurned
				}()
			}
		}(i)
	}

	go func() {
		for {
			select {
			case burned := <-ig.updateBurned:
				ig.burned = append(ig.burned, burned...)
			}
		}
	}()
}

// NewIncineratorGroup creates a new incinerator group from a number of
// incinerators. An incinerator group implements the same functionalities as
// an incinerator, so we can access them directly instead of viewing individual
// incinerators.
func NewIncineratorGroup(incinerators ...FullIncinerator) Incinerator {
	ig := &incineratorGroup{
		availableIncinerators: make(chan Incinerator, len(incinerators)),
		burned:                make([]Burner, 0),
		incinerators:          incinerators,
		updateBurned:          make(chan []Burner),
	}

	ig.loopBurn()
	return ig
}
