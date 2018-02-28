package goburnbooks

// SupplyPileGroup represents a group of SupplyPiles.
type SupplyPileGroup interface {
	SupplyPile
	Taken() []*SupplyTakeResult
}

type supplyPileGroup struct {
	supplyPiles  []SupplyPile
	taken        []*SupplyTakeResult
	takeResultCh chan *SupplyTakeResult
}

func (spg *supplyPileGroup) Supply(taker SupplyTaker) {
	for _, pile := range spg.supplyPiles {
		go pile.Supply(taker)
	}
}

func (spg *supplyPileGroup) Taken() []*SupplyTakeResult {
	return spg.taken
}

func (spg *supplyPileGroup) TakeResultChannel() <-chan *SupplyTakeResult {
	return spg.takeResultCh
}

// Loop supply to store available piles and take results.
func (spg *supplyPileGroup) loopSupply() {
	updateTaken := make(chan *SupplyTakeResult)

	for _, pile := range spg.supplyPiles {
		go func(pile SupplyPile) {
			takenResult := pile.TakeResultChannel()

			for {
				select {
				case result := <-takenResult:
					go func() {
						updateTaken <- result
					}()
				}
			}
		}(pile)
	}

	go func() {
		for {
			select {
			case taken := <-updateTaken:
				spg.taken = append(spg.taken, taken)

				go func() {
					spg.takeResultCh <- taken
				}()
			}
		}
	}()
}

// NewSupplyPileGroup creates a new SupplyPileGroup from a number of SupplyPiles.
func NewSupplyPileGroup(piles ...SupplyPile) SupplyPileGroup {
	group := &supplyPileGroup{
		supplyPiles: piles,
		taken:       make([]*SupplyTakeResult, 0),
	}

	go group.loopSupply()
	return group
}
