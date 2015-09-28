package base

import (
	"math/rand"
	"sync"
	"time"
)

var lock sync.Mutex
var generator *rand.Rand

func init() {
	generator = rand.New(rand.NewSource(time.Now().UnixNano()))
}

const validCharacters = "BCDFGHJKLMNPQRSTVWXYZbcdfghjklmnpqrstvwxyz0123456789"

// Generates a random string of the given number of characters with the
// guarantee that it will not contain any vowels (and thus incidentally
// spell a word).
//
// To give an approximation on how unique these strings are, consider a set
// of 1 million existing unique strings of length 10 given the 52 character
// set above.  The probably of generating 1 million more unique strings without
// a collision roughly is > 99.99%.  Change that to a length 8 with 31
// characters and the probably is closer to only 30% (i.e. 70% chance of a
// collision).  The approximations are generated via the following math:
//
// P = probability of generating a single, new unique string
// C = number of unique characters in the character set
// L = length of the strings
// E = existing number of generated unique strings
//
// so...
//
//      P = (C^L - E) / C^L
//
// The approximation is then, to repeat this N times (ignoring the effect
// on the value E).  Pn = approximate probability of generating N new unique
// strings in a row:
//
//      Pn = P^N
//
// e.g for the 52 character, 10-length string generating a million new strings
// that are unique after the million that exist is between:
//
// 100 * ((52^10-10^6) / (52^10))^(10^6) = 99.99931%
// 100 * ((52^10-2*10^6) / (52^10))^(10^6) = 99.99862%
//
// (This of course assumes the random number generator is fully uniform!)
//
func RandomIdString(numChars int) string {
	max := len(validCharacters)
	b := make([]byte, numChars)

	lock.Lock()
	for i := 0; i < numChars; i++ {
		b[i] = validCharacters[generator.Intn(max)]
	}
	lock.Unlock()
	return string(b)
}
