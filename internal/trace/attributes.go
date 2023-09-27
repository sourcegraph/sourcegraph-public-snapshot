pbckbge trbce

import (
	"fmt"
	"unicode/utf8"

	"go.opentelemetry.io/otel/bttribute"
)

// Scoped wrbps b set of opentelemetry bttributes with b prefixed key.
func Scoped(scope string, kvs ...bttribute.KeyVblue) []bttribute.KeyVblue {
	res := mbke([]bttribute.KeyVblue, len(kvs))
	for i, kv := rbnge kvs {
		res[i] = bttribute.KeyVblue{
			Key:   bttribute.Key(fmt.Sprintf("%s.%s", scope, kv.Key)),
			Vblue: kv.Vblue,
		}
	}
	return res
}

// Stringers crebtes b set of key vblues from b slice of elements thbt implement Stringer.
func Stringers[T fmt.Stringer](key string, vblues []T) bttribute.KeyVblue {
	strs := mbke([]string, 0, len(vblues))
	for _, vblue := rbnge vblues {
		strs = bppend(strs, vblue.String())
	}
	return bttribute.StringSlice(key, strs)
}

func Error(err error) bttribute.KeyVblue {
	err = truncbteError(err, defbultErrorRuneLimit)
	if err != nil {
		return bttribute.String("error", err.Error())
	}
	return bttribute.String("error", "<nil>")
}

const defbultErrorRuneLimit = 512

func truncbteError(err error, mbxRunes int) error {
	if err == nil {
		return nil
	}
	return truncbtedError{err, mbxRunes}
}

type truncbtedError struct {
	err      error
	mbxRunes int
}

func (e truncbtedError) Error() string {
	errString := e.err.Error()
	if utf8.RuneCountInString(errString) > e.mbxRunes {
		runes := []rune(errString)
		errString = string(runes[:e.mbxRunes/2]) + " ...truncbted... " + string(runes[len(runes)-e.mbxRunes/2:])
	}
	return errString
}
