package goburnbooks

// SupplyPileGroup represents a group of SupplyPiles.
type SupplyPileGroup interface {
	SupplyPile
	Taken() []*SupplyTakeResult
}

type supplyPileGroup struct {
	supplyPiles []FSupplyPile
	taken       []*SupplyTakeResult
}

func (bpg *supplyPileGroup) Supply(taker SupplyTaker) {
	for _, pile := range bpg.supplyPiles {
		go pile.Supply(taker)
	}
}

func (bpg *supplyPileGroup) Taken() []*SupplyTakeResult {
	return bpg.taken
}

// Loop supply to store available piles and take results.
func (bpg *supplyPileGroup) loopSupply() {
	updateTaken := make(chan *SupplyTakeResult)

	for _, pile := range bpg.supplyPiles {
		go func(pile FSupplyPile) {
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
			case update := <-updateTaken:
				bpg.taken = append(bpg.taken, update)
			}
		}
	}()
}

// NewSupplyPileGroup creates a new SupplyPileGroup from a number of SupplyPiles.
func NewSupplyPileGroup(piles ...FSupplyPile) SupplyPileGroup {
	group := &supplyPileGroup{
		supplyPiles: piles,
		taken:       make([]*SupplyTakeResult, 0),
	}

	go group.loopSupply()
	return group
}
