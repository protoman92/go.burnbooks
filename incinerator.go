package goburnbooks

// Incinerator represents something that can burn a Burnable.
type Incinerator interface {
	Consume(provider BurnableProvider)
}

// IncineratorParams represents the required parameters to set up an incinerator.
type IncineratorParams struct {
	Capacity int
	ID       string

	// This represents the minimum capacity required before this incinerator can
	// signal availability.
	MinCapacity int
}

// FIncinerator represents an incinerator that has all functionalities.
type FIncinerator interface {
	Incinerator
	BurnResult() <-chan *BurnResult
	UID() string
}

type incinerator struct {
	*IncineratorParams
	burnResult chan *BurnResult
}

func (i *incinerator) BurnResult() <-chan *BurnResult {
	return i.burnResult
}

func (i *incinerator) Consume(provider BurnableProvider) {
	go func() {
		capacity := i.Capacity
		burning := make(chan interface{}, capacity)
		burnResult := i.burnResult
		var provideCh = provider.ProvideChannel()
		var readyCh chan<- interface{}

		// Initialize this channel every time a new batch of Burnables is received.
		// Emissions from this channel means that enough items from a batch have
		// been processed.
		//
		// The sequence of the channel init is as such:
		// - A new batch is received, and this is initialized for that batch. Then
		// the provide channel is nullified.
		// - Once enough items have been burned, send a signal via this channel.
		// Then the provide channel is reinstated and this channel is nullified.
		//
		// Essentially these 2 channels are mutually exclusive.
		var enoughProcessedCh chan interface{}

		for {
			select {
			case burnables := <-provideCh:
				// Nullify the provide channel to let the sequence run in peace.
				provideCh = nil
				addProcessed := make(chan interface{})
				batchCount := len(burnables)
				enoughProcessedCh = make(chan interface{})
				processedCount := 0

				go func() {
					for {
						select {
						case <-addProcessed:
							processedCount++

							// Once we have processed enough items in a batch, send a signal
							// via the appropriate channel so that we can signal ready and
							// reinitialize the provide channel in order to receive the next
							// batch.
							if batchCount-processedCount < i.MinCapacity {
								addProcessed = nil
								enoughProcessedCh <- true
							}
						}
					}
				}()

				for _, burnable := range burnables {
					go func(burnable Burnable) {
						// Since this channel has a limited buffer, once the capacity is
						// reached this will block.
						burning <- true
						burnable.Burn()
						<-burning

						go func() {
							// This is in a goroutine because this channel could be nullified
							// halfway after enough items have been processed to prevent
							// duplicate signals.
							addProcessed <- true
						}()

						result := &BurnResult{Burned: burnable, IncineratorID: i.ID}
						burnResult <- result
					}(burnable)
				}

			case <-enoughProcessedCh:
				enoughProcessedCh = nil
				readyCh = provider.ConsumeReadyChannel()

				// Reinstate the provide channel to receive more requests.
				provideCh = provider.ProvideChannel()

			case readyCh <- true:
				readyCh = nil
			}
		}
	}()
}

func (i *incinerator) UID() string {
	return i.ID
}

// NewIncinerator creates a new incinerator with a specified pending channel
// and capacity. The capacity determines how many Burnables can be burned at any
// given point in time.
func NewIncinerator(params *IncineratorParams) FIncinerator {
	incinerator := &incinerator{
		IncineratorParams: params,
		burnResult:        make(chan *BurnResult),
	}

	return incinerator
}
