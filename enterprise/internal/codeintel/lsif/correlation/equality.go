package correlation

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func Equal(a, b *GroupedBundleDataMaps) bool {
	if !equalDocs(a, b) {
		return false
	}
	return true
}

func equalDocs(a, b *GroupedBundleDataMaps) bool {
	if len(a.Documents) != len(b.Documents) {
		return false
	}

	for path, aDoc := range a.Documents {
		bDoc, exists := b.Documents[path]
		if !exists {
			return false
		}

		if len(aDoc.Ranges) != len(bDoc.Ranges) {
			return false
		}

		aRngIDs := sortedRangeIDs(aDoc.Ranges)
		bRngIDs := sortedRangeIDs(bDoc.Ranges)
		for idx, aID := range aRngIDs {
			aRng := aDoc.Ranges[aID]
			bRng := bDoc.Ranges[bRngIDs[idx]]

			if lsifstore.CompareRanges(aRng, bRng) != 0 {
				return false
			}

			aResult := Resolve(a, aDoc, aRng)
			bResult := Resolve(b, bDoc, bRng)

			if !equalResult(aResult, bResult) {
				return false
			}
		}
	}

	return true
}

func equalResult(a, b QueryResult) bool {
	return true
}
