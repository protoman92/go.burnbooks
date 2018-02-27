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
	Cap     int
	ID      string
	LoadCh  chan<- []Suppliable
	ReadyCh chan interface{}
}

type supplyTaker struct {
	*SupplyTakerParams
}

func (bt *supplyTaker) Capacity() int {
	return bt.Cap
}

func (bt *supplyTaker) LoadChannel() chan<- []Suppliable {
	return bt.LoadCh
}

func (bt *supplyTaker) ReadyChannel() <-chan interface{} {
	return bt.ReadyCh
}

func (bt *supplyTaker) UID() string {
	return bt.ID
}

// NewSupplyTaker creates a new SupplyTaker.
func NewSupplyTaker(params *SupplyTakerParams) SupplyTaker {
	return &supplyTaker{SupplyTakerParams: params}
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
