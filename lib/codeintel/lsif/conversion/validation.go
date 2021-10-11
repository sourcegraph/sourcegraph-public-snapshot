package conversion

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion/datastructures"
)

func validate(state *State) error {
	rangeToDoc := map[int]int{}
	var err error
	state.Contains.Each(func(doc1 int, ranges *datastructures.IDSet) {
		if err != nil {
			return
		}

		ranges.Each(func(r int) {
			if err != nil {
				return
			}

			if doc2, ok := rangeToDoc[r]; ok {
				err = fmt.Errorf("validate: range %d is contained in multiple documents (at least %d and %d, possibly more)", r, doc1, doc2)
				return
			} else {
				rangeToDoc[r] = doc1
			}
		})
	})
	if err != nil {
		return err
	}

	nameToDataMap := map[string]map[int]*datastructures.DefaultIDSetMap{
		"DefinitionData": state.DefinitionData,
		"ReferenceData":  state.ReferenceData,
	}

	for name, dataMap := range nameToDataMap {
		for _, m := range dataMap {
			m.Each(func(doc int, ranges *datastructures.IDSet) {
				if err != nil {
					return
				}

				ranges.Each(func(r int) {
					if err != nil {
						return
					}

					if doc2, ok := rangeToDoc[r]; ok {
						if doc != doc2 {
							err = fmt.Errorf("validate: range %d is contained in document %d, but linked to a different document %d by %s", r, doc2, doc, name)
						}
					} else {
						err = fmt.Errorf("validate: range %d from DefinitionData isn't contained by any documents", r)
					}
				})
			})
		}
	}

	return err
}
