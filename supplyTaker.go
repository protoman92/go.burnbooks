package goburnbooks

import (
	"fmt"
	"time"
)

// SupplyTaker represents a worker that takes Suppliables for some purposes.
type SupplyTaker interface {
	Capacity() uint
	SupplyTakerID() string

	// This channel receives supplies from supply piles.
	ReceiveLoadChannel() chan<- []Suppliable

	// This channel sends ready signal to supply piles.
	SendTakeReadyChannel() <-chan interface{}
}

// SupplyTakerRawParams represents only the immutable parameters used to build
// a taker.
type SupplyTakerRawParams struct {
	Cap         uint
	STID        string
	TakeTimeout time.Duration
}

// SupplyTakerParams represents all the required parameters to build a taker.
type SupplyTakerParams struct {
	SupplyTakerRawParams
	SendSupplyDestCh chan<- []Suppliable
	STLogger         Logger
}

type supplyTaker struct {
	SupplyTakerParams
	receiveLoadCh   chan []Suppliable
	sendTakeReadyCh chan interface{}
}

func (st *supplyTaker) String() string {
	return fmt.Sprintf("Supply taker %s", st.STID)
}

func (st *supplyTaker) Capacity() uint {
	return st.Cap
}

func (st *supplyTaker) ReceiveLoadChannel() chan<- []Suppliable {
	return st.receiveLoadCh
}

func (st *supplyTaker) SendTakeReadyChannel() <-chan interface{} {
	return st.sendTakeReadyCh
}

func (st *supplyTaker) SupplyTakerID() string {
	return st.STID
}

func (st *supplyTaker) loopWork() {
	logger := st.STLogger
	sendTakeReadyCh := st.sendTakeReadyCh
	resetSequenceCh := make(chan interface{}, 1)
	var receiveLoadCh chan []Suppliable
	var sendSupplyDestCh chan<- []Suppliable
	var suppliables []Suppliable
	var takeTimeoutCh <-chan time.Time

	for {
		select {
		case sendTakeReadyCh <- true:
			sendTakeReadyCh = nil
			receiveLoadCh = st.receiveLoadCh
			takeTimeoutCh = time.After(st.TakeTimeout)

		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		case suppliables = <-receiveLoadCh:
			logger.Printf("%v received %d supplies from source", st, len(suppliables))
			receiveLoadCh = nil
			takeTimeoutCh = nil
			sendSupplyDestCh = st.SendSupplyDestCh

		case <-takeTimeoutCh:
			logger.Printf("%v timed out!", st)
			takeTimeoutCh = nil
			receiveLoadCh = nil
			resetSequenceCh <- true
		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

		case sendSupplyDestCh <- suppliables:
			logger.Printf("%v sent %d supplies to destination", st, len(suppliables))
			sendSupplyDestCh = nil
			suppliables = nil
			resetSequenceCh <- true

		case <-resetSequenceCh:
			sendTakeReadyCh = st.sendTakeReadyCh
		}
	}
}

// NewSupplyTaker creates a new SupplyTaker.
func NewSupplyTaker(params *SupplyTakerParams) SupplyTaker {
	supplyTaker := &supplyTaker{
		SupplyTakerParams: *params,
		receiveLoadCh:     make(chan []Suppliable),
		sendTakeReadyCh:   make(chan interface{}),
	}

	go supplyTaker.loopWork()
	return supplyTaker
}
