package privacy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCombinePrivacy(t *testing.T) {
	testCases := []struct {
		p1     Privacy
		p2     Privacy
		result Privacy
	}{
		{Private, Private, Private},
		{Unknown, Private, Private},
		{Unknown, Unknown, Unknown},
		{Public, Private, Private},
		{Public, Unknown, Unknown},
		{Public, Public, Public},
	}
	for _, testCase := range testCases {
		require.Equal(t, testCase.result, testCase.p1.Combine(testCase.p2))
		require.Equal(t, testCase.result, testCase.p2.Combine(testCase.p1))
	}
}
