/* Copyright 2017 The Bazel Authors. All rights reserved.

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

// Package repo provides functionality for managing Go repository rules.
//
// UNSTABLE: The exported APIs in this package may change. In the future,
// language extensions should implement an interface for repository
// rule management. The update-repos command will call interface methods,
// and most if this package's functionality will move to language/go.
// Moving this package to an internal directory would break existing
// extensions, since RemoteCache is referenced through the resolve.Resolver
// interface, which extensions are required to implement.
package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/rule"
)

const gazelleFromDirectiveKey = "_gazelle_from_directive"

// FindExternalRepo attempts to locate the directory where Bazel has fetched
// the external repository with the given name. An error is returned if the
// repository directory cannot be located.
func FindExternalRepo(repoRoot, name string) (string, error) {
	// See https://docs.bazel.build/versions/master/output_directories.html
	// for documentation on Bazel directory layout.
	// We expect the bazel-out symlink in the workspace root directory to point to
	// <output-base>/execroot/<workspace-name>/bazel-out
	// We expect the external repository to be checked out at
	// <output-base>/external/<name>
	externalPath := strings.Join([]string{repoRoot, "bazel-out", "..", "..", "..", "external", name}, string(os.PathSeparator))
	cleanPath, err := filepath.EvalSymlinks(externalPath)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(cleanPath)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("%s: not a directory", externalPath)
	}
	return cleanPath, nil
}

type macroKey struct {
	file, def string
}

type loader struct {
	repos        []*rule.Rule
	repoRoot     string
	repoFileMap  map[string]*rule.File // repo rule name => file that contains repo
	repoIndexMap map[string]int        // repo rule name => index of rule in "repos" slice
	visited      map[macroKey]struct{}
}

// IsFromDirective returns true if the repo rule was defined from a directive.
func IsFromDirective(repo *rule.Rule) bool {
	b, ok := repo.PrivateAttr(gazelleFromDirectiveKey).(bool)
	return ok && b
}

// add adds a repository rule to a file.
// In the case of duplicate rules, select the rule
// with the following prioritization:
//   - rules that were provided as directives have precedence
//   - rules that were provided last
func (l *loader) add(file *rule.File, repo *rule.Rule) {
	name := repo.Name()
	if name == "" {
		return
	}

	if i, ok := l.repoIndexMap[repo.Name()]; ok {
		if IsFromDirective(l.repos[i]) && !IsFromDirective(repo) {
			// We always prefer directives over non-directives
			return
		}
		// Update existing rule to new rule
		l.repos[i] = repo
	} else {
		l.repos = append(l.repos, repo)
		l.repoIndexMap[name] = len(l.repos) - 1
	}
	l.repoFileMap[name] = file
}

// visit returns true exactly once for each file,function key, and false for all future instances
func (l *loader) visit(file, function string) bool {
	if _, ok := l.visited[macroKey{file, function}]; ok {
		return false
	}
	l.visited[macroKey{file, function}] = struct{}{}
	return true
}

// ListRepositories extracts metadata about repositories declared in a
// file.
func ListRepositories(workspace *rule.File) (repos []*rule.Rule, repoFileMap map[string]*rule.File, err error) {
	l := &loader{
		repoRoot:     filepath.Dir(workspace.Path),
		repoIndexMap: make(map[string]int),
		repoFileMap:  make(map[string]*rule.File),
		visited:      make(map[macroKey]struct{}),
	}

	for _, repo := range workspace.Rules {
		l.add(workspace, repo)
	}
	if err := l.loadExtraRepos(workspace); err != nil {
		return nil, nil, err
	}

	for _, d := range workspace.Directives {
		switch d.Key {
		case "repository_macro":
			parsed, err := ParseRepositoryMacroDirective(d.Value)
			if err != nil {
				return nil, nil, err
			}

			if err := l.loadRepositoriesFromMacro(parsed); err != nil {
				return nil, nil, err
			}
		}
	}
	return l.repos, l.repoFileMap, nil
}

func (l *loader) loadRepositoriesFromMacro(macro *RepoMacro) error {
	f := filepath.Join(l.repoRoot, macro.Path)
	if !l.visit(f, macro.DefName) {
		return nil
	}

	macroFile, err := rule.LoadMacroFile(f, "", macro.DefName)
	if err != nil {
		return fmt.Errorf("failed to load %s in repoRoot %s: %w", f, l.repoRoot, err)
	}
	loads := map[string]*rule.Load{}
	for _, load := range macroFile.Loads {
		for _, name := range load.Symbols() {
			loads[name] = load
		}
	}
	for _, rule := range macroFile.Rules {
		// (Incorrectly) assume that anything with a name attribute is a rule, not a macro to recurse into
		if rule.Name() != "" {
			l.add(macroFile, rule)
			continue
		}
		if !macro.Leveled {
			continue
		}
		// If another repository macro is loaded that macro defName must be called.
		// When a defName is called, the defName of the function is the rule's "kind".
		// This then must be matched with the Load that it is imported with, so that
		// file can be loaded
		kind := rule.Kind()
		load := loads[kind]
		if load == nil {
			continue
		}
		resolved := loadToMacroDef(load, l.repoRoot, kind)
		// TODO: Also handle the case where one macro calls another macro in the same bzl file
		if macro.Path == "" {
			continue
		}

		if err := l.loadRepositoriesFromMacro(resolved); err != nil {
			return err
		}
	}
	return l.loadExtraRepos(macroFile)
}

// loadToMacroDef takes a load
// e.g. for if called on
// load("package_name:package_dir/file.bzl", alias_name="original_def_name")
// with defAlias = "alias_name", it will return:
//
//	-> "/Path/to/package_name/package_dir/file.bzl"
//	-> "original_def_name"
func loadToMacroDef(l *rule.Load, repoRoot, defAlias string) *RepoMacro {
	rel := strings.Replace(filepath.Clean(l.Name()), ":", string(filepath.Separator), 1)
	// A loaded macro may refer to the macro by a different name (alias) in the load,
	// thus, the original name must be resolved to load the macro file properly.
	defName := l.Unalias(defAlias)
	return &RepoMacro{
		Path:    rel,
		DefName: defName,
	}
}

func (l *loader) loadExtraRepos(f *rule.File) error {
	extraRepos, err := parseRepositoryDirectives(f.Directives)
	if err != nil {
		return err
	}
	for _, repo := range extraRepos {
		l.add(f, repo)
	}
	return nil
}

func parseRepositoryDirectives(directives []rule.Directive) (repos []*rule.Rule, err error) {
	for _, d := range directives {
		switch d.Key {
		case "repository":
			vals := strings.Fields(d.Value)
			if len(vals) < 2 {
				return nil, fmt.Errorf("failure parsing repository: %s, expected repository kind and attributes", d.Value)
			}
			kind := vals[0]
			r := rule.NewRule(kind, "")
			r.SetPrivateAttr(gazelleFromDirectiveKey, true)
			for _, val := range vals[1:] {
				kv := strings.SplitN(val, "=", 2)
				if len(kv) != 2 {
					return nil, fmt.Errorf("failure parsing repository: %s, expected format for attributes is attr1_name=attr1_value", d.Value)
				}
				r.SetAttr(kv[0], kv[1])
			}
			if r.Name() == "" {
				return nil, fmt.Errorf("failure parsing repository: %s, expected a name attribute for the given repository", d.Value)
			}
			repos = append(repos, r)
		}
	}
	return repos, nil
}

type RepoMacro struct {
	Path    string
	DefName string
	Leveled bool
}

// ParseRepositoryMacroDirective checks the directive is in proper format, and splits
// path and defName. Repository_macros prepended with a "+" (e.g. "# gazelle:repository_macro +file%def")
// indicates a "leveled" macro, which loads other macro files.
func ParseRepositoryMacroDirective(directive string) (*RepoMacro, error) {
	vals := strings.Split(directive, "%")
	if len(vals) != 2 {
		return nil, fmt.Errorf("Failure parsing repository_macro: %s, expected format is macroFile%%defName", directive)
	}
	f := vals[0]
	if strings.HasPrefix(f, "..") {
		return nil, fmt.Errorf("Failure parsing repository_macro: %s, macro file path %s should not start with \"..\"", directive, f)
	}
	return &RepoMacro{
		Path:    strings.TrimPrefix(f, "+"),
		DefName: vals[1],
		Leveled: strings.HasPrefix(f, "+"),
	}, nil
}
