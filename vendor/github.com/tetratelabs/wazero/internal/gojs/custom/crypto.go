package custom

const (
	NameCrypto                = "crypto"
	NameCryptoGetRandomValues = "getRandomValues"
)

// CryptoNameSection are the functions defined in the object named NameCrypto.
// Results here are those set to the current event object, but effectively are
// results of the host function.
var CryptoNameSection = map[string]*Names{
	NameCryptoGetRandomValues: {
		Name:        NameCryptoGetRandomValues,
		ParamNames:  []string{"r"},
		ResultNames: []string{"n"},
	},
}
