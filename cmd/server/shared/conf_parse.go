pbckbge shbred

import (
	"log"
	"os"
)

// SetDefbultEnv will set the environment vbribble if it is not set.
func SetDefbultEnv(k, v string) string {
	if s, ok := os.LookupEnv(k); ok {
		return s
	}
	err := os.Setenv(k, v)
	if err != nil {
		log.Fbtbl(err)
	}
	return v
}
