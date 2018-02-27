package goburnbooks

import (
	"fmt"
)

// SupplyTaker represents a worker that takes Suppliables for some purposes.
type SupplyTaker interface {
	Capacity() int
	LoadChannel() chan<- []Suppliable
	ReadyChannel() <-chan interface{}
	UID() string
}

// SupplyTakerParams represents all the required parameters to build a taker.
type SupplyTakerParams struct {
	capacity int
	id       string
	loadCh   chan<- []Suppliable
	readyCh  chan interface{}
}

type supplyTaker struct {
	*SupplyTakerParams
}

func (bt *supplyTaker) Capacity() int {
	return bt.capacity
}

func (bt *supplyTaker) LoadChannel() chan<- []Suppliable {
	return bt.loadCh
}

func (bt *supplyTaker) ReadyChannel() <-chan interface{} {
	return bt.readyCh
}

func (bt *supplyTaker) UID() string {
	return bt.id
}

// NewSupplyTaker creates a new SupplyTaker.
func NewSupplyTaker(params *SupplyTakerParams) SupplyTaker {
	return &supplyTaker{SupplyTakerParams: params}
}

// SupplyTakeResult represents the result of a take operation.
type SupplyTakeResult struct {
	pileID    string
	takerID   string
	supplyIds []string
}

func (btr *SupplyTakeResult) String() string {
	return fmt.Sprintf(
		"Supply taker %s took %d supplies from pile %s",
		btr.takerID,
		len(btr.supplyIds),
		btr.pileID,
	)
}
