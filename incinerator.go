package goburnbooks

// Incinerator represents something that can burn a Burnable.
type Incinerator interface {
	// Although we can feed Burnables to the pending channel directly, calling
	// this method allows us to define custom behavior, such as only allowing
	// those with capacity to accept more Burnables.
	Incinerate(burnables ...Burnable)
}

// IncineratorParams represents the required parameters to set up an incinerator.
type IncineratorParams struct {
	Capacity int

	// This represents the minimum capacity required before this incinerator can
	// signal availability.
	MinCapacity int

	UID string
}

// FIncinerator represents an incinerator that has all functionalities, such
// as signalling availability.
type FIncinerator interface {
	Available
	BurnResultCollector
	Incinerator
	Pending() chan<- Burnable
}

type incinerator struct {
	IncineratorParams
	available  chan interface{}
	burnResult chan *BurnResult
	pending    chan Burnable
}

func (i *incinerator) String() string {
	return i.UID
}

func (i *incinerator) SignalAvailable() <-chan interface{} {
	return i.available
}

func (i *incinerator) BurnResult() <-chan *BurnResult {
	return i.burnResult
}

func (i *incinerator) Incinerate(burnables ...Burnable) {
	go func() {
		for _, burnable := range burnables {
			i.pending <- burnable
		}
	}()
}

func (i *incinerator) Pending() chan<- Burnable {
	return i.pending
}

func (i *incinerator) loopBurn() {
	burning := make(chan interface{}, i.Capacity)
	burningCount := 0
	updateBurning := make(chan int)

	for {
		select {
		case update := <-updateBurning:
			burningCount += update

			// The number of Burnables in pending pile may be way more than spare
			// capacity, so here we enforce that only when available slots is more
			// than a minimum do we signal availability.
			if i.Capacity-burningCount > i.MinCapacity {
				go func() {
					i.available <- true
				}()
			}

		case burnable := <-i.pending:
			go func() {
				// If the incinerator capacity is reached, this should block.
				burning <- true

				// If this statement is not in a goroutine, it will block because it
				// is not buffered. As a result, we will never reach the update code.
				updateBurning <- 1
				burnable.Burn()

				// Release a slot for the next Burnable.
				<-burning

				// Similar reasoning to above as to why this statement has to be in
				// a goroutine.
				updateBurning <- -1
				result := &BurnResult{incineratorID: i.UID, burned: burnable}
				i.burnResult <- result
			}()
		}
	}
}

// NewIncinerator creates a new incinerator with a specified pending channel
// and capacity. The capacity determines how many Burnables can be burned at any
// given point in time.
//
// The pending channel is provided here because we do not want to impose how
// many Burnables are allowed to pile up before workers have to wait. For e.g.,
// there may be a stack of 100 books in front of the incinerator because a
// worker feels like carrying that many at once.
func NewIncinerator(params IncineratorParams, pending chan Burnable) FIncinerator {
	incinerator := &incinerator{
		IncineratorParams: params,
		available:         make(chan interface{}),
		burnResult:        make(chan *BurnResult),
		pending:           pending,
	}

	go func() {
		incinerator.available <- true
	}()

	go incinerator.loopBurn()
	return incinerator
}
