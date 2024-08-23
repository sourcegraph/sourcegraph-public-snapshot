// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctags

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type CTagsParserType uint8

const (
	UnknownCTags CTagsParserType = iota
	NoCTags
	UniversalCTags
	ScipCTags
)

const debug = false

type LanguageMap = map[string]CTagsParserType

func ParserToString(parser CTagsParserType) string {
	switch parser {
	case UnknownCTags:
		return "unknown"
	case NoCTags:
		return "no"
	case UniversalCTags:
		return "universal"
	case ScipCTags:
		return "scip"
	default:
		panic("Reached impossible CTagsParserType state")
	}
}

func StringToParser(str string) CTagsParserType {
	switch str {
	case "no":
		return NoCTags
	case "universal":
		return UniversalCTags
	case "scip":
		return ScipCTags
	default:
		return UniversalCTags
	}
}

type ParserBinMap map[CTagsParserType]string

func NewParserBinMap(
	ctagsPath string,
	scipCTagsPath string,
	languageMap LanguageMap,
	cTagsMustSucceed bool,
) (ParserBinMap, error) {
	validBins := make(map[CTagsParserType]string)
	requiredBins := map[CTagsParserType]string{UniversalCTags: ctagsPath}
	for _, parserType := range languageMap {
		if parserType == ScipCTags {
			requiredBins[ScipCTags] = scipCTagsPath
			break
		}
	}

	for parserType, bin := range requiredBins {
		if bin == "" && cTagsMustSucceed {
			return nil, fmt.Errorf("ctags binary not found for %s parser type", ParserToString(parserType))
		}
		if err := checkBinary(parserType, bin); err != nil && cTagsMustSucceed {
			return nil, fmt.Errorf("ctags.NewParserBinMap: %v", err)
		}
		validBins[parserType] = bin
	}

	return validBins, nil
}

// checkBinary does checks on bin to ensure we can correctly use the binary
// for symbols. It is more user friendly to fail early in this case.
func checkBinary(typ CTagsParserType, bin string) error {
	switch typ {
	case UniversalCTags:
		helpOutput, err := exec.Command(bin, "--help").CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to check if %s is universal-ctags: %w\n--help output:\n%s", bin, err, string(helpOutput))
		}
		if !bytes.Contains(helpOutput, []byte("+interactive")) {
			return fmt.Errorf("ctags binary is not universal-ctags or is not compiled with +interactive feature: bin=%s", bin)
		}

	case ScipCTags:
		if !strings.Contains(bin, "scip-ctags") {
			return fmt.Errorf("only supports scip-ctags, not %s", bin)
		}
	}

	return nil
}
