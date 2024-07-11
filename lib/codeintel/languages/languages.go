package languages

import (
	"slices"
	"strings"

	"github.com/go-enry/go-enry/v2" //nolint:depguard - This package is allowed to use enry.
)

// Make sure all names are lowercase here, since they are normalized
var enryLanguageMappings = map[string]string{
	"c#": "c_sharp",
}

func NormalizeLanguage(filetype string) string {
	normalized := strings.ToLower(filetype)
	if mapped, ok := enryLanguageMappings[normalized]; ok {
		normalized = mapped
	}

	return normalized
}

// GetMostLikelyLanguage returns the language for the given path and contents.
//
// Prefer using GetLanguages instead of this function.
//
// TODO: Remove the extra normalization this functiond does over GetLanguages
func GetMostLikelyLanguage(path, contents string) (lang string, found bool) {
	languages, _ := GetLanguages(path, func() ([]byte, error) {
		if len(contents) > 2048 {
			return []byte(contents[:2048]), nil
		}
		return []byte(contents), nil
	})
	for _, lang := range languages {
		if lang != "" {
			return NormalizeLanguage(lang), true
		}
	}
	return "", false
}

// GetLanguages is a replacement for enry.GetLanguages which
// avoids incorrect fallback behavior that is present in DefaultStrategies,
// where it will misclassify '.h' header files as C when file contents
// are not available.
//
// The content can be optionally passed via a callback instead of
// directly, so that in the common case, the caller can avoid fetching
// the content. The full content returned by getContent will be used for
// language detection.
//
// getContent is not called if the file is likely to be a binary file,
// as enry only covers programming languages.
//
// Returns:
//   - An error if the getContent func returns an error
//   - An empty slice if language detection failed
//   - A single-element slice if the language was determined exactly
//   - A multi-element slice if the language was ambiguous. For example,
//     for simple `.h` files with just comments and macros, they may
//     be valid C, C++ or any of their derivative languages (e.g. Objective-C).
func GetLanguages(path string, getContent func() ([]byte, error)) ([]string, error) {
	impl := func() ([]string, error) {
		langs := enry.GetLanguagesByFilename(path, nil, nil)
		if len(langs) == 1 {
			return langs, nil
		}
		newLangs, isLikelyBinaryFile := getLanguagesByExtension(path)
		if isLikelyBinaryFile {
			return nil, nil
		}
		switch len(newLangs) {
		case 0:
			break
		case 1:
			return newLangs, nil
		default:
			langs = newLangs
		}
		if getContent == nil {
			return langs, nil
		}
		content, err := getContent()
		if err != nil {
			return nil, err
		}
		if len(content) == 0 {
			return langs, nil
		}
		if enry.IsBinary(content) {
			return nil, nil
		}

		// enry doesn't expose a way to call GetLanguages with a specific set of
		// strategies, so just hand-roll that code here.
		var languages = langs
		for _, strategy := range []enry.Strategy{enry.GetLanguagesByModeline, getLanguagesByShebang, enry.GetLanguagesByContent, enry.GetLanguagesByClassifier} {
			candidates := strategy(path, content, languages)
			switch len(candidates) {
			case 0:
				continue
			case 1:
				return candidates, nil
			default:
				languages = candidates
			}
		}

		return languages, nil
	}

	langs, err := impl()
	return slices.Clone(langs), err
}

// getLanguagesByShebang is a replacement for enry.GetLanguagesByShebang.
//
// The enry function considers non-programming languages such as 'Pod'/'Pod 6'
// also for shebangs, so work around that.
func getLanguagesByShebang(path string, content []byte, candidates []string) []string {
	languages := enry.GetLanguagesByShebang(path, content, candidates)
	if len(languages) == 2 {
		// See https://sourcegraph.com/github.com/go-enry/go-enry@40f2a1e5b90eec55c20441c2a5911dcfc298a447/-/blob/data/interpreter.go?L95-96
		if slices.Equal(languages, []string{"Perl", "Pod"}) {
			return []string{"Perl"}
		}
		if slices.Equal(languages, []string{"Pod 6", "Raku"}) {
			return []string{"Raku"}
		}
	}
	return slices.Clone(languages)
}
