package goburnbooks

import (
	"fmt"
)

// BurnableProvider represents a Burnable provider.
type BurnableProvider interface {
	BurnableProviderID() string

	// This channel receives ready signals from incinerators. Only when these
	// signals are received do we send burnables.
	//
	// Beware that an emission does not mean the previous work load has been
	// finished, just that the incinerator has burned enough to take in more, and
	// it does not expect the remaining load to take long.
	ReceiveProvideReadyChannel() chan<- string

	// This channel sends burnables to incinerators.
	SendBurnablesChannel() <-chan []Burnable
}

// BurnableProviderRawParams represents only the immutable parameters used to
// build a provider.
type BurnableProviderRawParams struct {
	BPID string
}

// BurnableProviderParams represents all the required parameters to build a
// provider.
type BurnableProviderParams struct {
	BurnableProviderRawParams
	BPLogger                Logger
	ReceiveBurnableSourceCh <-chan []Burnable
}

// The already providing channel prevents multiple incinerators from sending
// ready signals to this provider.
type burnableProvider struct {
	BurnableProviderParams
	receiveProvideReadyCh chan string
	sendBurnablesCh       chan []Burnable
}

func (bp *burnableProvider) String() string {
	return fmt.Sprintf("Provider %s", bp.BPID)
}

func (bp *burnableProvider) SendBurnablesChannel() <-chan []Burnable {
	return bp.sendBurnablesCh
}

func (bp *burnableProvider) ReceiveProvideReadyChannel() chan<- string {
	return bp.receiveProvideReadyCh
}

func (bp *burnableProvider) BurnableProviderID() string {
	return bp.BPID
}

func (bp *burnableProvider) loopWork() {
	logger := bp.BPLogger
	receiveProvideReadyCh := bp.receiveProvideReadyCh
	var burnables []Burnable
	var receiveBurnablesCh <-chan []Burnable
	var sendBurnablesCh chan []Burnable

	for {
		select {
		case incID := <-receiveProvideReadyCh:
			logger.Printf("%v received ready signal from incinerator %v", bp, incID)
			receiveProvideReadyCh = nil
			receiveBurnablesCh = bp.ReceiveBurnableSourceCh

		case burnables = <-receiveBurnablesCh:
			receiveBurnablesCh = nil
			sendBurnablesCh = bp.sendBurnablesCh

		case sendBurnablesCh <- burnables:
			burnables = nil
			sendBurnablesCh = nil
			receiveProvideReadyCh = bp.receiveProvideReadyCh
		}
	}
}

// NewBurnableProvider returns a new BurnableProvider.
func NewBurnableProvider(params *BurnableProviderParams) BurnableProvider {
	bp := &burnableProvider{
		BurnableProviderParams: *params,
		receiveProvideReadyCh:  make(chan string),
		sendBurnablesCh:        make(chan []Burnable),
	}

	go bp.loopWork()
	return bp
}
