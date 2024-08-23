/* Copyright 2018 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resolve

import (
	"flag"
	"log"
	"regexp"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// FindRuleWithOverride searches the current configuration for user-specified
// dependency resolution overrides. Overrides specified later (in configuration
// files in deeper directories, or closer to the end of the file) are
// returned first. If no override is found, label.NoLabel is returned.
func FindRuleWithOverride(c *config.Config, imp ImportSpec, lang string) (label.Label, bool) {
	rc := getResolveConfig(c)
	if dep, ok := rc.findOverride(imp, lang); ok {
		return dep, true
	}
	for i := len(rc.regexpOverrides) - 1; i >= 0; i-- {
		o := rc.regexpOverrides[i]
		if o.matches(imp, lang) {
			return o.dep, true
		}
	}
	return label.NoLabel, false
}

type overrideKey struct {
	imp  ImportSpec
	lang string
}

type regexpOverrideSpec struct {
	ImpLang  string
	ImpRegex *regexp.Regexp
	lang     string
	dep      label.Label
}

func (o regexpOverrideSpec) matches(imp ImportSpec, lang string) bool {
	return imp.Lang == o.ImpLang &&
		o.ImpRegex.MatchString(imp.Imp) &&
		(o.lang == "" || o.lang == lang)
}

type resolveConfig struct {
	overrides       map[overrideKey]label.Label
	regexpOverrides []regexpOverrideSpec
	parent          *resolveConfig
}

// newResolveConfig creates a new resolveConfig with the given overrides and
// regexpOverrides. If the new overrides are the same as the parent's, the
// parent is returned instead.
func newResolveConfig(parent *resolveConfig, newOverrides map[overrideKey]label.Label, regexpOverrides []regexpOverrideSpec) *resolveConfig {
	if len(newOverrides) == 0 && len(regexpOverrides) == len(parent.regexpOverrides) {
		return parent
	}
	return &resolveConfig{
		overrides:       newOverrides,
		regexpOverrides: regexpOverrides,
		parent:          parent,
	}
}

// findOverride searches the current configuration for an override matching
// the given import and language. If no override is found, the parent
// configuration is searched recursively.
func (rc *resolveConfig) findOverride(imp ImportSpec, lang string) (label.Label, bool) {
	key := overrideKey{imp: imp, lang: lang}
	if dep, ok := rc.overrides[key]; ok {
		return dep, ok
	}
	if rc.parent != nil {
		return rc.parent.findOverride(imp, lang)
	}
	return label.NoLabel, false
}

const resolveName = "_resolve"

func getResolveConfig(c *config.Config) *resolveConfig {
	return c.Exts[resolveName].(*resolveConfig)
}

type Configurer struct{}

func (*Configurer) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	c.Exts[resolveName] = &resolveConfig{}
}

func (*Configurer) CheckFlags(fs *flag.FlagSet, c *config.Config) error { return nil }

func (*Configurer) KnownDirectives() []string {
	return []string{"resolve", "resolve_regexp"}
}

func (*Configurer) Configure(c *config.Config, rel string, f *rule.File) {
	if f == nil || len(f.Directives) == 0 {
		return
	}

	rc := getResolveConfig(c)
	var newOverrides map[overrideKey]label.Label
	regexpOverrides := rc.regexpOverrides[:len(rc.regexpOverrides):len(rc.regexpOverrides)]

	for _, d := range f.Directives {
		if d.Key == "resolve" {
			parts := strings.Fields(d.Value)
			key := overrideKey{}
			var lbl string
			if len(parts) == 3 {
				key.imp.Lang = parts[0]
				key.lang = parts[0]
				key.imp.Imp = parts[1]
				lbl = parts[2]
			} else if len(parts) == 4 {
				key.imp.Lang = parts[0]
				key.lang = parts[1]
				key.imp.Imp = parts[2]
				lbl = parts[3]
			} else {
				log.Printf("could not parse directive: %s\n\texpected gazelle:resolve source-language [import-language] import-string label", d.Value)
				continue
			}
			dep, err := label.Parse(lbl)
			if err != nil {
				log.Printf("gazelle:resolve %s: %v", d.Value, err)
				continue
			}
			dep = dep.Abs("", rel)
			if newOverrides == nil {
				newOverrides = make(map[overrideKey]label.Label, len(f.Directives))
			}
			newOverrides[key] = dep
		} else if d.Key == "resolve_regexp" {
			parts := strings.Fields(d.Value)
			o := regexpOverrideSpec{}
			var lbl string
			if len(parts) == 3 {
				o.ImpLang = parts[0]
				var err error
				o.ImpRegex, err = regexp.Compile(parts[1])
				if err != nil {
					log.Printf("gazelle:resolve_regexp %s: %v", d.Value, err)
					continue
				}
				lbl = parts[2]
			} else if len(parts) == 4 {
				o.ImpLang = parts[0]
				o.lang = parts[1]
				var err error
				o.ImpRegex, err = regexp.Compile(parts[2])
				if err != nil {
					log.Printf("gazelle:resolve_regexp %s: %v", d.Value, err)
					continue
				}

				lbl = parts[3]
			} else {
				log.Printf("could not parse directive: %s\n\texpected gazelle:resolve_regexp source-language [import-language] import-string-regex label", d.Value)
				continue
			}
			var err error
			o.dep, err = label.Parse(lbl)
			if err != nil {
				log.Printf("gazelle:resolve_regexp %s: %v", d.Value, err)
				continue
			}
			o.dep = o.dep.Abs("", rel)
			regexpOverrides = append(regexpOverrides, o)
		}
	}

	c.Exts[resolveName] = newResolveConfig(rc, newOverrides, regexpOverrides)
}
