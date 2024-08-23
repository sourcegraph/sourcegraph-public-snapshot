/* Copyright 2023 The Bazel Authors. All rights reserved.

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

package language

import (
	"flag"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// BaseLang implements the minimum of language.Language interface.
// This is not meant to be used directly by Gazelle, but to be used by
// a downstream struct through composition. End users could use this to
// write an extensions iteratively without having to implement every
// functions in the interface right away.
//
// Example usage:
//
//	type MyLang struct {
//		language.BaseLang
//	}
//
//	func NewLanguage() language.Language {
//		return &MyLang{}
//	}
type BaseLang struct{}

func (b *BaseLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}

func (b *BaseLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	return nil
}

func (b *BaseLang) KnownDirectives() []string {
	return nil
}

func (b *BaseLang) Configure(c *config.Config, rel string, f *rule.File) {}

func (b *BaseLang) Name() string {
	return "BaseLang"
}

func (b *BaseLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	return nil
}

func (b *BaseLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

func (b *BaseLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
}

func (b *BaseLang) Kinds() map[string]rule.KindInfo {
	return nil
}

func (b *BaseLang) Loads() []rule.LoadInfo {
	return nil
}

func (b *BaseLang) GenerateRules(args GenerateArgs) GenerateResult {
	return GenerateResult{}
}

func (b *BaseLang) Fix(c *config.Config, f *rule.File) {}
