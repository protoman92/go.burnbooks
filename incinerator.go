package goburnbooks

// Burner represents something that can burn, e.g. books. In the Burn() method,
// we may implement variable sleep durations to simulate different burning
// processes (bigger books burn more slowly).
//
// For the sake of simplicity, we assume that everything can be burnt eventually,
// only that some do so longer than others.
type Burner interface {
	Burn()
}

// Incinerator represents something that can burn a Burner.
type Incinerator interface {
	Burned() []Burner
}

type incinerator struct {
	burning      chan Burner
	burned       []Burner
	pending      <-chan Burner
	updateBurned chan Burner
}

func (i *incinerator) Burned() []Burner {
	return i.burned
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
func NewIncinerator(capacity int, pending <-chan Burner) Incinerator {
	incinerator := &incinerator{
		burned:       make([]Burner, 0),
		burning:      make(chan Burner, capacity),
		pending:      pending,
		updateBurned: make(chan Burner),
	}

	go incinerator.loopBurn()
	return incinerator
}

// IncineratorGroup represents a group of incinerators.
type IncineratorGroup interface{}

type incineratorGroup struct {
	incinerators []Incinerator
}

// NewIncineratorGroup creates a new incinerator group from a number of
// incinerators.
func NewIncineratorGroup(incinerators ...Incinerator) IncineratorGroup {
	ig := &incineratorGroup{incinerators: incinerators}
	return ig
}
