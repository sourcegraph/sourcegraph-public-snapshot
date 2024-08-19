// This file wraps the logic of go-enry (https://github.com/go-enry/go-enry) to support additional languages.
// go-enry is based off of a package called Linguist (https://github.com/github/linguist)
// and sometimes programming languages may not be supported by Linguist
// or may take a while to get merged in and make it into go-enry. This wrapper
// gives us flexibility to support languages in those cases. We list additional languages
// in this file and remove them once they make it into Linguist and go-enry.
// This logic is similar to what we have in the sourcegraph/sourcegraph repo, in the future
// we plan to refactor both into a common library to share between the two repos.
package languages

import (
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

var unsupportedByLinguistAliasMap = map[string]string{
	// Extensions for the Apex programming language
	// See https://developer.salesforce.com/docs/atlas.en-us.apexcode.meta/apexcode/apex_dev_guide.htm
	"apex": "Apex",
	// Pkl Configuration Language (https://pkl-lang.org/)
	// Add to linguist on 6/7/24
	// can remove once go-enry package updates
	// to that linguist version
	"pkl": "Pkl",
	// Magik Language
	"magik": "Magik",
}

var unsupportedByLinguistExtensionToNameMap = map[string]string{
	".apex":    "Apex",
	".apxt":    "Apex",
	".apxc":    "Apex",
	".cls":     "Apex",
	".trigger": "Apex",
	// Pkl Configuration Language (https://pkl-lang.org/)
	".pkl": "Pkl",
	// Magik Language
	".magik": "Magik",
}

// getLanguagesByAlias is a replacement for enry.GetLanguagesByAlias
// It supports languages that are missing in linguist
func GetLanguageByAlias(alias string) (language string, ok bool) {
	language, ok = enry.GetLanguageByAlias(alias)
	if !ok {
		normalizedAlias := strings.ToLower(alias)
		language, ok = unsupportedByLinguistAliasMap[normalizedAlias]
	}

	return
}

// GetLanguage is a replacement for enry.GetLanguage
// to find out the most probable language to return but includes support
// for languages missing from linguist
func GetLanguage(filename string, content []byte) (language string) {
	language = enry.GetLanguage(filename, content)

	// If go-enry failed to find language, fall back on our
	// internal check for languages missing in linguist
	if language == "" {
		ext := filepath.Ext(filename)
		normalizedExt := strings.ToLower(ext)
		if ext == "" {
			return
		}
		if lang, ok := unsupportedByLinguistExtensionToNameMap[normalizedExt]; ok {
			language = lang
		}
	}
	return
}
