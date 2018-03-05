package goburnbooks

import (
	"fmt"
	"sync"
)

// Incinerator represents something that can burn a Burnable.
type Incinerator interface {
	BurnResultChannel() <-chan BurnResult
	Consume(provider BurnableProvider)
	UID() string
}

// FIncinerator represents an incinerator that has all functionalities.
type FIncinerator interface {
	Incinerator
}

// IncineratorParams represents the required parameters to set up an incinerator.
type IncineratorParams struct {
	Capacity uint
	Logger   Logger
	ID       string

	// This represents the minimum capacity required before this incinerator can
	// signal availability.
	MinCapacity uint
}

// The consume ready channel is here to coordinate access to the incinerator
// by allowing only one provider to provide burnables at any time. Thus, it has
// a buffer of 1.
type incinerator struct {
	IncineratorParams
	burnResultCh chan BurnResult
}

func (i *incinerator) String() string {
	return fmt.Sprintf("Incinerator %s", i.ID)
}

func (i *incinerator) BurnResultChannel() <-chan BurnResult {
	return i.burnResultCh
}

func (i *incinerator) Consume(provider BurnableProvider) {
	go func() {
		capacity := i.Capacity
		burnResult := i.burnResultCh
		burning := make(chan interface{}, capacity)
		logger := i.Logger
		providerID := provider.BurnableProviderID()
		provideReadyCh := provider.ReceiveProvideReadyChannel()
		resetSequenceCh := make(chan interface{}, 1)
		var provideCh <-chan []Burnable

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
			case provideReadyCh <- i.ID:
				logger.Printf("%v is ready to consume from %v", i, provider)
				provideReadyCh = nil
				provideCh = provider.SendBurnablesChannel()

			case burnables := <-provideCh:
				// Nullify the provide channel to let the sequence run in peace.
				logger.Printf("%v received %d from %v", i, len(burnables), provider)
				provideCh = nil
				batchCount := uint(len(burnables))
				enoughProcessedCh = make(chan interface{}, 1)

				if batchCount == 0 {
					enoughProcessedCh <- true
					break
				}

				addProcessed := make(chan interface{})
				processedCount := uint(0)

				var mutex sync.RWMutex

				accessAddProcessed := func() chan interface{} {
					mutex.RLock()
					defer mutex.RUnlock()
					return addProcessed
				}

				nullifyAddProcessed := func() {
					mutex.Lock()
					defer mutex.Unlock()
					addProcessed = nil
				}

				go func() {
					for {
						select {
						case <-accessAddProcessed():
							processedCount++

							// Once we have processed enough items in a batch, send a signal
							// via the appropriate channel so that we can signal ready and
							// reinitialize the provide channel in order to receive the next
							// batch.
							if batchCount-processedCount < i.MinCapacity {
								enoughProcessedCh <- true
								nullifyAddProcessed()
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
							if addProcessed := accessAddProcessed(); addProcessed != nil {
								addProcessed <- true
							}
						}()

						result := NewBurnResult(burnable, i.ID, providerID)
						burnResult <- result
					}(burnable)
				}

			case <-enoughProcessedCh:
				logger.Printf("%v has burned enough, signalling ready to %v", i, provider)
				enoughProcessedCh = nil
				resetSequenceCh <- true

			case <-resetSequenceCh:
				provideReadyCh = provider.ReceiveProvideReadyChannel()
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
	i := &incinerator{
		IncineratorParams: *params,
		burnResultCh:      make(chan BurnResult),
	}

	if i.Capacity < i.MinCapacity {
		panic(fmt.Sprintf(
			"%v has capacity %d less than min capacity %d",
			i,
			i.Capacity,
			i.MinCapacity,
		))
	}

	return i
}
