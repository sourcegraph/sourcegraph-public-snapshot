package languages

import (
	"path/filepath"
	"slices"

	"github.com/go-enry/go-enry/v2" //nolint:depguard - Only this package can use enry
)

// GetLanguageByNameOrAlias returns the standardized name for
// a language based on its name (in which case this is an identity operation)
// or based on its alias, which is potentially an alternate name for
// the language.
//
// Aliases are fully lowercase, and map N-1 to languages.
//
// For example,
//
// GetLanguageByNameOrAlias("ada") == "Ada", true
// GetLanguageByNameOrAlias("ada95") == "Ada", true
//
// Historical note: This function was added for replacing usages of
// enry.GetLanguageByAlias, which, unlike the name suggests, also
// handles non-normalized names such as those with spaces.
func GetLanguageByNameOrAlias(nameOrAlias string) (lang string, ok bool) {
	alias := convertToAliasKey(nameOrAlias)
	if lang, ok = unsupportedByEnryAliasMap[alias]; ok {
		return lang, true
	}

	return enry.GetLanguageByAlias(alias)
}

// GetLanguageExtensions returns the list of file extensions for a given
// language. Returned extensions are always prefixed with a '.'.
//
// The returned slice will be empty iff the language is not known.
//
// Handles more languages than enry.GetLanguageExtensions.
//
// Mutually consistent with getLanguagesByExtension, see the tests
// for the exact invariants.
func GetLanguageExtensions(language string) []string {
	if langs, ok := unsupportedByEnryNameToExtensionMap[language]; ok {
		return langs
	}

	ignoreExts, isNiche := nicheExtensionUsages[language]
	// Force a copy to avoid accidentally modifying the global variable
	enryExts := slices.Clone(enry.GetLanguageExtensions(language))
	if !isNiche {
		return slices.Clone(enryExts)
	}
	return slices.DeleteFunc(enryExts, func(ext string) bool {
		_, shouldIgnore := ignoreExts[ext]
		return shouldIgnore
	})
}

// getLanguagesByExtension is a replacement for enry.GetLanguagesByExtension
// to work around the following limitations:
//   - For some extensions which are overwhelmingly used by a certain file type
//     in practice, such as '.ts', '.md' and '.yaml', it returns ambiguous results.
//   - It does not provide any information about binary files.
//   - Some languages are not supported by enry yet (e.g. Pkl)
func getLanguagesByExtension(path string) (candidates []string, isLikelyBinaryFile bool) {
	ext := filepath.Ext(path)
	if ext == "" {
		return nil, false
	}
	if lang, ok := unsupportedByEnryExtensionToNameMap[ext]; ok {
		return []string{lang}, false
	}
	if _, ok := commonBinaryFileExtensions[ext[1:]]; ok {
		return nil, true
	}
	if lang, ok := overrideAmbiguousExtensionsMap[ext]; ok {
		return []string{lang}, false
	}
	return enry.GetLanguagesByExtension(path, nil, nil), false
}

var commonBinaryFileExtensions = func() map[string]struct{} {
	m := map[string]struct{}{}
	for _, s := range commonBinaryFileExtensionsList {
		m[s] = struct{}{}
	}
	return m
}()

var overrideAmbiguousExtensionsMap = map[string]string{
	// Ignoring the uncommon usage of '.cs' for Smalltalk.
	".cs": "C#",
	// The other languages are Filterscript, Forth, GLSL. Out of that,
	// Forth and GLSL commonly use other extensions. Ignore Filterscript
	// as it is niche.
	".fs": "F#",
	// Ignoring other variants of JSON, such as OASv2-json and OASv3-json
	".json": "JSON",
	// Not considering "GCC Machine Description".
	".md": "Markdown",
	// The other main language using '.rs' is RenderScript, but that's deprecated.
	// See https://developer.android.com/guide/topics/renderscript/compute
	".rs": "Rust",
	// In i18n contexts, there are XML files with '.ts' and '.tsx' extensions,
	// but we ignore those for now to avoid penalizing the common case.
	".tsx": "TSX",
	".ts":  "TypeScript",
	// Ignoring "Adblock Filter List" and "Vim Help File".
	".txt": "Text",
	// Ignoring other variants of YAML, such as MiniYAML, OASv2-yaml, OASv3-yaml.
	".yaml": "YAML",
	".yml":  "YAML",
}

var unsupportedByEnryExtensionToNameMap = map[string]string{
	// Extensions for the Apex programming language
	// See https://developer.salesforce.com/docs/atlas.en-us.apexcode.meta/apexcode/apex_dev_guide.htm
	".apex":    "Apex",
	".apxt":    "Apex",
	".apxc":    "Apex",
	".cls":     "Apex",
	".trigger": "Apex",
	// See TODO(id: remove-pkl-special-case)
	".pkl":   "Pkl",
	".magik": "Magik",
}

// nicheExtensionUsage keeps track of which (lang, extension) mappings
// should not be considered.
//
// We cannot wholesale ignore these languages, as this list includes
// languages like XML, but it can contain unusual extensions like '.tsx'
// which we generally want to classify as TypeScript.
var nicheExtensionUsages = func() map[string]map[string]struct{} {
	niche := map[string]map[string]struct{}{}
	considered := map[string]struct{}{}
	for _, lang := range overrideAmbiguousExtensionsMap {
		considered[lang] = struct{}{}
	}
	for ext := range overrideAmbiguousExtensionsMap {
		langs := enry.GetLanguagesByExtension("foo"+ext, nil, nil)
		for _, lang := range langs {
			if _, found := considered[lang]; !found {
				if m, hasMap := niche[lang]; hasMap {
					m[ext] = struct{}{}
				} else {
					niche[lang] = map[string]struct{}{ext: {}}
				}
			}
		}
	}
	for specialOverrideExt, lang := range unsupportedByEnryExtensionToNameMap {
		considered[lang] = struct{}{}
		langs := enry.GetLanguagesByExtension("foo"+specialOverrideExt, nil, nil)
		for _, lang := range langs {
			if _, found := considered[lang]; !found {
				if m, hasMap := niche[lang]; hasMap {
					m[specialOverrideExt] = struct{}{}
				} else {
					niche[lang] = map[string]struct{}{specialOverrideExt: {}}
				}
			}
		}
	}
	return niche
}()

var unsupportedByEnryNameToExtensionMap = reverseMap(unsupportedByEnryExtensionToNameMap)

// unsupportedByEnryAliasMap maps alias -> language name for languages
// not tracked by go-enry.
var unsupportedByEnryAliasMap = func() map[string]string {
	out := map[string]string{}
	for _, lang := range unsupportedByEnryExtensionToNameMap {
		out[convertToAliasKey(lang)] = lang
	}
	return out
}()

func reverseMap(m map[string]string) map[string][]string {
	n := make(map[string][]string, len(m))
	for k, v := range m {
		n[v] = append(n[v], k)
	}
	return n
}

// Source: https://github.com/sindresorhus/binary-extensions/blob/main/binary-extensions.json
// License: https://github.com/sindresorhus/binary-extensions/blob/main/license
// Replace the contents with
// curl -L https://raw.githubusercontent.com/sindresorhus/binary-extensions/main/binary-extensions.json | jq '.[]' | awk  '{print $1 ","}'
//
// Not adding a leading '.' here to make it easier to update/compare the list.
var commonBinaryFileExtensionsList = []string{
	"3dm",
	"3ds",
	"3g2",
	"3gp",
	"7z",
	"a",
	"aac",
	"adp",
	"ai",
	"aif",
	"aiff",
	"alz",
	"ape",
	"apk",
	"appimage",
	"ar",
	"arj",
	"asf",
	"au",
	"avi",
	"bak",
	"baml",
	"bh",
	"bin",
	"bk",
	"bmp",
	"btif",
	"bz2",
	"bzip2",
	"cab",
	"caf",
	"cgm",
	"class",
	"cmx",
	"cpio",
	"cr2",
	"cur",
	"dat",
	"dcm",
	"deb",
	"dex",
	"djvu",
	"dll",
	"dmg",
	"dng",
	"doc",
	"docm",
	"docx",
	"dot",
	"dotm",
	"dra",
	"DS_Store",
	"dsk",
	"dts",
	"dtshd",
	"dvb",
	"dwg",
	"dxf",
	"ecelp4800",
	"ecelp7470",
	"ecelp9600",
	"egg",
	"eol",
	"eot",
	"epub",
	"exe",
	"f4v",
	"fbs",
	"fh",
	"fla",
	"flac",
	"flatpak",
	"fli",
	"flv",
	"fpx",
	"fst",
	"fvt",
	"g3",
	"gh",
	"gif",
	"graffle",
	"gz",
	"gzip",
	"h261",
	"h263",
	"h264",
	"icns",
	"ico",
	"ief",
	"img",
	"ipa",
	"iso",
	"jar",
	"jpeg",
	"jpg",
	"jpgv",
	"jpm",
	"jxr",
	"key",
	"ktx",
	"lha",
	"lib",
	"lvp",
	"lz",
	"lzh",
	"lzma",
	"lzo",
	"m3u",
	"m4a",
	"m4v",
	"mar",
	"mdi",
	"mht",
	"mid",
	"midi",
	"mj2",
	"mka",
	"mkv",
	"mmr",
	"mng",
	"mobi",
	"mov",
	"movie",
	"mp3",
	"mp4",
	"mp4a",
	"mpeg",
	"mpg",
	"mpga",
	"mxu",
	"nef",
	"npx",
	"numbers",
	"nupkg",
	"o",
	"odp",
	"ods",
	"odt",
	"oga",
	"ogg",
	"ogv",
	"otf",
	"ott",
	"pages",
	"pbm",
	"pcx",
	"pdb",
	"pdf",
	"pea",
	"pgm",
	"pic",
	"png",
	"pnm",
	"pot",
	"potm",
	"potx",
	"ppa",
	"ppam",
	"ppm",
	"pps",
	"ppsm",
	"ppsx",
	"ppt",
	"pptm",
	"pptx",
	"psd",
	"pya",
	"pyc",
	"pyo",
	"pyv",
	"qt",
	"rar",
	"ras",
	"raw",
	"resources",
	"rgb",
	"rip",
	"rlc",
	"rmf",
	"rmvb",
	"rpm",
	"rtf",
	"rz",
	"s3m",
	"s7z",
	"scpt",
	"sgi",
	"shar",
	"snap",
	"sil",
	"sketch",
	"slk",
	"smv",
	"snk",
	"so",
	"stl",
	"suo",
	"sub",
	"swf",
	"tar",
	"tbz",
	"tbz2",
	"tga",
	"tgz",
	"thmx",
	"tif",
	"tiff",
	"tlz",
	"ttc",
	"ttf",
	"txz",
	"udf",
	"uvh",
	"uvi",
	"uvm",
	"uvp",
	"uvs",
	"uvu",
	"viv",
	"vob",
	"war",
	"wav",
	"wax",
	"wbmp",
	"wdp",
	"weba",
	"webm",
	"webp",
	"whl",
	"wim",
	"wm",
	"wma",
	"wmv",
	"wmx",
	"woff",
	"woff2",
	"wrm",
	"wvx",
	"xbm",
	"xif",
	"xla",
	"xlam",
	"xls",
	"xlsb",
	"xlsm",
	"xlsx",
	"xlt",
	"xltm",
	"xltx",
	"xm",
	"xmind",
	"xpi",
	"xpm",
	"xwd",
	"xz",
	"z",
	"zip",
	"zipx",
}
