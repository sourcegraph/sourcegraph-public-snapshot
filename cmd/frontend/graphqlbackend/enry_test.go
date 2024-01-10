package graphqlbackend

import (
	"testing"

	"github.com/go-enry/go-enry/v2"
	"github.com/stretchr/testify/require"
)

var matlabFile string = `% matlab function to compute square of a value
	function [out] = square(x)
		out = x * x;
	end

	function [out] = fourthpower(x)
		out = square(square(x));
	end`

func TestEnryLangs(t *testing.T) {
	langs := enry.GetLanguages("foo.m", []byte(matlabFile))
	require.Equal(t, []string{"MATLAB"}, langs)
}
