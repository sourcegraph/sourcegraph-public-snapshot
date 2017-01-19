package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"

type dependencyReferencesResolver struct {
	deps     []*dependencyReferenceResolver
	location *locationResolver
}

type locationResolver struct {
	location *lspLocationResolver
	symbol   map[string]interface{}
}

type lspLocationResolver struct {
	uri            string
	startLine      int32
	startCharacter int32
	endLine        int32
	endCharacter   int32
}

type dependencyReferenceResolver struct {
	dep *sourcegraph.DependencyReference
}

func (r *dependencyReferencesResolver) Deps() []*dependencyReferenceResolver {
	return r.deps
}

func (r *dependencyReferencesResolver) Location() *locationResolver {
	return r.location
}

func (r *dependencyReferenceResolver) DepData() map[string]interface{} {
	return r.dep.DepData
}

func (r *dependencyReferenceResolver) RepoID() int32 {
	return r.dep.RepoID
}

func (r *dependencyReferenceResolver) Hints() map[string]interface{} {
	return r.dep.Hints
}

func (r *locationResolver) Location() *lspLocationResolver {
	return r.location
}

func (r *locationResolver) Symbol() map[string]interface{} {
	return r.symbol
}

func (r *lspLocationResolver) URI() string {
	return r.uri
}

func (r *lspLocationResolver) StartLine() int32 {
	return r.startLine
}

func (r *lspLocationResolver) StartCharacter() int32 {
	return r.startCharacter
}

func (r *lspLocationResolver) EndLine() int32 {
	return r.endLine
}

func (r *lspLocationResolver) EndCharacter() int32 {
	return r.endCharacter
}
