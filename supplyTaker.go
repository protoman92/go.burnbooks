package goburnbooks

import (
	"fmt"
)

// SupplyTaker represents a worker that takes Suppliables for some purposes.
type SupplyTaker interface {
	Capacity() uint
	LoadChannel() chan<- []Suppliable
	SupplyTakerID() string
	TakeReadyChannel() <-chan interface{}
}

// SupplyTakerRawParams represents only the immutable parameters used to build
// a taker.
type SupplyTakerRawParams struct {
	Cap  uint
	STID string
}

// SupplyTakerParams represents all the required parameters to build a taker.
type SupplyTakerParams struct {
	SupplyTakerRawParams
	LoadCh      chan<- []Suppliable
	TakeReadyCh chan interface{}
}

type supplyTaker struct {
	SupplyTakerParams
}

func (bt *supplyTaker) Capacity() uint {
	return bt.Cap
}

func (bt *supplyTaker) LoadChannel() chan<- []Suppliable {
	return bt.LoadCh
}

func (bt *supplyTaker) TakeReadyChannel() <-chan interface{} {
	return bt.TakeReadyCh
}

func (bt *supplyTaker) SupplyTakerID() string {
	return bt.STID
}

// NewSupplyTaker creates a new SupplyTaker.
func NewSupplyTaker(params *SupplyTakerParams) SupplyTaker {
	return &supplyTaker{SupplyTakerParams: *params}
}

// SupplyTakeResult represents the result of a take operation.
type SupplyTakeResult struct {
	PileID    string
	TakerID   string
	SupplyIds []string
}

func (btr *SupplyTakeResult) String() string {
	return fmt.Sprintf(
		"Supply taker %s took %d supplies from pile %s",
		btr.TakerID,
		len(btr.SupplyIds),
		btr.PileID,
	)
}
