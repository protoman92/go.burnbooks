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
	Burned() <-chan Burner
}

type incinerator struct {
	burning chan Burner
	burned  chan Burner
	pending <-chan Burner
}

func (i *incinerator) Burned() <-chan Burner {
	return i.burned
}

func (i *incinerator) loopBurn() {
	for {
		select {
		case burner := <-i.burning:
			go func() {
				burner.Burn()
				i.burned <- burner
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
func NewIncinerator(capacity int, pending <-chan Burner) Incinerator {
	burned := make(chan Burner)
	burning := make(chan Burner, capacity)

	incinerator := &incinerator{
		burned:  burned,
		burning: burning,
		pending: pending,
	}

	go incinerator.loopBurn()
	return incinerator
}
