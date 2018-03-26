// +build !go1.7,!go1.8

package gocode

func init() {
	knownPackageIdents["context"] = "golang.org/x/net/context"
}
