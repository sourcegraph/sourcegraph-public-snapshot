// Package dep defines an interface for listing raw dependencies and resolving
// them, and registering handlers to perform these tasks.
//
// Toolchains should register two handlers to list and resolve dependencies:
//
// *Lister:* this performs a quick pass over the declared dependencies of a
// source unit and emits whatever information can be determined without
// an expensive dependency resolution process. For example, the Go lister
// simply lists the packages imported by each file in the source unit.
//
// *Resolver:* this resolves a RawDependency (emitted from a Lister),
// determining the defining repository, source unit, version, etc. For example,
// the Ruby resolver queries rubygems.org (and consults a hard-coded list) to
// determine this information. Note that the only necessary field in
// ResolvedTarget is ToRepoCloneURL; the other fields are for future use.
//
// In code, this looks like:
//
//   package my_toolchain
//
//   func init() {
//     dep.RegisterLister(PythonPackage{}, pip)
//     dep.RegisterResolver("pip-package", pip)
//   }
//
//   type pip struct {}
//
//   func (p *pip) List(dir string, unit unit.SourceUnit, c *config.Repository) ([]*dep.RawDependency, error) {
//     // return each line in requirements.txt, perhaps
//     // return something like:
//     return []*dep.RawDependency{{TargetType: "pip-package", Target: "Django==1.6"}}, nil
//   }
//
//   func (p *pip) Resolve(rawDep *dep.RawDependency, c *config.Repository) (*dep.ResolvedTarget, error) {
//     // query pypi.python.org or use cheerio, and then return something like:
//     return &dep.ResolvedTarget{
//       ToRepoCloneURL: "git://github.com/django/django.git",
//     }, nil
//   }
//
// Separating the list and resolution steps allows us to more easily pare down
// the total number of unique resolutions we must perform. If multiple source
// units all depend on the same external library, then those duplicate
// RawDependencies only yield a single call to Resolve. This saves even more
// work when finding dependencies for all of a repository's history (because
// most commits likely won't add or update dependency declarations).
//
// TODO(sqs): update these docs after we removed deplist
package dep
