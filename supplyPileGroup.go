package goburnbooks

import (
	"sync"
)

// SupplyPileGroup represents a group of SupplyPiles.
type SupplyPileGroup interface {
	SupplyPile
	SupplyPileContribMap() map[string]int
	SupplyTakerContribMap() map[string]int
	Taken() []SupplyTakeResult
}

type supplyPileGroup struct {
	mutex       sync.RWMutex
	supplyPiles []FSupplyPile
	taken       []SupplyTakeResult
}

func (spg *supplyPileGroup) Supply(taker SupplyTaker) {
	for _, pile := range spg.supplyPiles {
		go pile.Supply(taker)
	}
}

func (spg *supplyPileGroup) SupplyPileContribMap() map[string]int {
	spg.mutex.RLock()
	taken := spg.taken
	spg.mutex.RUnlock()

	contributorMap := make(map[string]int, 0)

	for _, taken := range taken {
		id := taken.PileID()
		contributorMap[id] = contributorMap[id] + len(taken.SupplyIDs())
	}

	return contributorMap
}

func (spg *supplyPileGroup) SupplyTakerContribMap() map[string]int {
	spg.mutex.RLock()
	taken := spg.taken
	spg.mutex.RUnlock()

	contributorMap := make(map[string]int, 0)

	for _, taken := range taken {
		id := taken.TakerID()
		contributorMap[id] = contributorMap[id] + len(taken.SupplyIDs())
	}

	return contributorMap
}

func (spg *supplyPileGroup) Taken() []SupplyTakeResult {
	spg.mutex.RLock()
	defer spg.mutex.RUnlock()
	return spg.taken
}

// Loop supply to store available piles and take results.
func (spg *supplyPileGroup) loopSupply() {
	for _, pile := range spg.supplyPiles {
		go func(pile FSupplyPile) {
			for {
				result, ok := <-pile.TakeResultChannel()

				if ok {
					// Note that this mutex is only used to modify the result map, since
					// said map will be accessible via a getter method.
					spg.mutex.Lock()
					spg.taken = append(spg.taken, result)
					spg.mutex.Unlock()
				} else {
					return
				}
			}
		}(pile)
	}
}

// NewSupplyPileGroup creates a new SupplyPileGroup from a number of SupplyPiles.
func NewSupplyPileGroup(piles ...FSupplyPile) SupplyPileGroup {
	group := &supplyPileGroup{
		supplyPiles: piles,
		taken:       make([]SupplyTakeResult, 0),
	}

	go group.loopSupply()
	return group
}
