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
	BurnResultCollector
	Incinerator
	Availability() <-chan chan<- []Burnable
}

type incinerator struct {
	*IncineratorParams
	available  chan chan<- []Burnable
	burnResult chan *BurnResult
	pending    chan []Burnable
}

func (i *incinerator) Availability() <-chan chan<- []Burnable {
	return i.available
}

func (i *incinerator) BurnResult() <-chan *BurnResult {
	return i.burnResult
}

func (i *incinerator) Incinerate(burnables ...Burnable) {
	go func() {
		i.pending <- burnables
	}()
}

func (i *incinerator) loopBurn() {
	burning := make(chan interface{}, i.Capacity)
	availableCount := i.Capacity
	updateAvailable := make(chan int)

	for {
		select {
		case update := <-updateAvailable:
			// We serialize count updates with a channel. Admittedly this is the not
			// the best approach because sometimes the count can get below 0, but it
			// works well enough to limit the number of times this incinerator signals
			// availability.
			availableCount += update

			// The number of Burnables in pending pile may be way more than spare
			// capacity, so here we enforce that only when available slots is more
			// than a minimum do we signal availability.
			if availableCount > i.MinCapacity {
				go func() {
					// In this implementation, an incinerator will receive a whole load
					// when it signals availability, and the load will be kept pending
					// instead of being directed elsewhere in case not all Burnables can
					// be processed.
					//
					// If this were a real system, the rationale would be to minimize
					// message count related to re-distributing workload.
					i.available <- i.pending
				}()
			}

		case burnables := <-i.pending:
			for _, burnable := range burnables {
				go func(burnable Burnable) {
					// If the incinerator capacity is reached, this should block.
					burning <- true

					// If this statement is not in a goroutine, it will block because it
					// is not buffered. As a result, we will never reach the update code.
					updateAvailable <- -1
					burnable.Burn()

					// Release a slot for the next Burnable.
					<-burning

					// Similar reasoning to above as to why this statement has to be in
					// a goroutine.
					updateAvailable <- 1
					result := &BurnResult{incineratorID: i.UID, burned: burnable}
					i.burnResult <- result
				}(burnable)
			}
		}
	}
}

// NewIncinerator creates a new incinerator with a specified pending channel
// and capacity. The capacity determines how many Burnables can be burned at any
// given point in time.
func NewIncinerator(params *IncineratorParams) FIncinerator {
	incinerator := &incinerator{
		IncineratorParams: params,
		available:         make(chan chan<- []Burnable),
		burnResult:        make(chan *BurnResult),
		pending:           make(chan []Burnable),
	}

	go func() {
		incinerator.available <- incinerator.pending
	}()

	go incinerator.loopBurn()
	return incinerator
}
