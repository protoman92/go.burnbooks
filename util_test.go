package goburnbooks_test

import gbb "goburnbooks"

func burnedIDMap(ig gbb.IncineratorGroup) map[string]int {
	allBurned := ig.Burned()
	burnedMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.Burned.BurnableID()
		burnedMap[id] = burnedMap[id] + 1
	}

	return burnedMap
}

func burnContributorMap(ig gbb.IncineratorGroup) map[string]int {
	allBurned := ig.Burned()
	contributorMap := make(map[string]int, 0)

	for _, burned := range allBurned {
		id := burned.IncineratorID
		contributorMap[id] = contributorMap[id] + 1
	}

	return contributorMap
}
