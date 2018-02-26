package goburnbooks

// Burner represents something that can burn, e.g. books. In the Burn() method,
// we may implement variable sleep durations to simulate different burning
// processes (bigger books burn more slowly).
//
// For the sake of simplicity, we assume that everything can be burnt eventually,
// only that some do so longer than others. Therefore, the Burn() method does
// not error out.
type Burner interface {
	Burn()
}

// Incinerator represents something that can burn a Burner.
type Incinerator interface {
	Burned() []Burner

	// Although we can feed Burners to the pending channel directly, calling this
	// method allows us to define custom behavior, such as only allowing those
	// with capacity to accept more Burner.
	Feed(burners ...Burner)
}

// FullIncinerator represents an incinerator that has all functionalities. such
// as signalling availability.
type FullIncinerator interface {
	Available
	Incinerator
}

type incinerator struct {
	available    chan bool
	burning      chan Burner
	burned       []Burner
	pending      chan Burner
	updateBurned chan Burner
}

func (i *incinerator) Available() <-chan bool {
	return i.available
}

func (i *incinerator) Burned() []Burner {
	return i.burned
}

func (i *incinerator) Feed(burners ...Burner) {
	go func() {
		for _, burner := range burners {
			i.pending <- burner
		}
	}()
}

func (i *incinerator) loopBurn() {
	for {
		select {
		case burner := <-i.burning:
			go func() {
				burner.Burn()
				i.updateBurned <- burner
			}()

		case burned := <-i.updateBurned:
			// We update burned books here to serialize slice update.
			i.burned = append(i.burned, burned)

			go func() {
				i.available <- true
			}()

		case burner := <-i.pending:
			go func() {
				i.burning <- burner
			}()
		}
	}
}

// NewIncinerator creates a new incinerator with a specified pending channel
// and capacity. The capacity determines how many Burners can be burned at any
// given point in time.
//
// The pending channel is provided here because we do not want to impose how
// many Burners are allowed to pile up before workers have to wait. For e.g.,
// there may be a stack of 100 books in front of the incinerator because a
// worker feels like carrying that many at once.
func NewIncinerator(capacity int, pending chan Burner) FullIncinerator {
	incinerator := &incinerator{
		available:    make(chan bool),
		burned:       make([]Burner, 0),
		burning:      make(chan Burner, capacity),
		pending:      pending,
		updateBurned: make(chan Burner),
	}

	go func() {
		incinerator.available <- true
	}()

	go incinerator.loopBurn()
	return incinerator
}
