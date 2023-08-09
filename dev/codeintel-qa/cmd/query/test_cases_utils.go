package main

import "fmt"

// Location specifies the first position in a source range.
type Location struct {
	Repo      string
	Rev       string
	Path      string
	Line      int
	Character int
}

type TaggedLocation struct {
	Location                   Location
	IgnoreSiblingRelationships bool
}

const maxRefToDefAssertionsPerFile = 10

// generate tests that asserts definition <> reference relationships on a particular set of
// locations all referring to the same SCIP symbol
func makeDefsRefsTests(symbolName string, defs []Location, refs []TaggedLocation) (fns []queryFunc) {
	var untagagedRefs []Location
	for _, taggedLocation := range refs {
		untagagedRefs = append(untagagedRefs, taggedLocation.Location)
	}

	for _, def := range defs {
		fns = append(fns,
			makeDefsTest(symbolName, "definition", def, defs),          // "you are at definition"
			makeRefsTest(symbolName, "definition", def, untagagedRefs), // def -> refs
		)
	}

	sourceFiles := map[string]int{}

	for _, ref := range refs {
		if ref.IgnoreSiblingRelationships {
			continue
		}

		sourceFiles[ref.Location.Path] = sourceFiles[ref.Location.Path] + 1
		if sourceFiles[ref.Location.Path] >= maxRefToDefAssertionsPerFile {
			continue
		}

		// ref -> def
		fns = append(fns, makeDefsTest(symbolName, "reference", ref.Location, defs))

		if queryReferencesOfReferences {
			// global search for other refs
			fns = append(fns, makeRefsTest(symbolName, "reference", ref.Location, untagagedRefs))
		}
	}

	return fns
}

// generate tests that asserts prototype <> implementation relationships on a particular set of
// locations all referring to the same SCIP symbol
func makeProtoImplsTests(symbolName string, prototype Location, implementations []Location) (fns []queryFunc) {
	fns = append(fns,
		// N.B.: unlike defs/refs tests, prototypes don't "implement" themselves so we do not
		// assert that prototypes of a prototype is an identity function (unlike def -> def).
		makeImplsTest(symbolName, "prototype", prototype, implementations),
	)

	for _, implementation := range implementations {
		fns = append(fns,
			// N.B.: unlike defs/refs tests, sibling implementations do not "implement" each other
			// so we do not assert implementations can jump to siblings without first going to the
			// prototype.
			makeProtosTest(symbolName, "implementation", implementation, []Location{prototype}),
		)
	}

	return fns
}

// generate tests that asserts the definitions at the given source location
func makeDefsTest(symbolName, target string, source Location, expectedResults []Location) queryFunc {
	return makeTestFunc(fmt.Sprintf("definitions of %s from %s", symbolName, target), queryDefinitions, source, expectedResults)
}

// generate tests that asserts the references at the given source location
func makeRefsTest(symbolName, target string, source Location, expectedResults []Location) queryFunc {
	return makeTestFunc(fmt.Sprintf("references of %s from %s", symbolName, target), queryReferences, source, expectedResults)
}

// generate tests that asserts the prototypes at the given source location
func makeProtosTest(symbolName, target string, source Location, expectedResults []Location) queryFunc {
	return makeTestFunc(fmt.Sprintf("prototypes of %s from %s", symbolName, target), queryPrototypes, source, expectedResults)
}

// generate tests that asserts the implementations at the given source location
func makeImplsTest(symbolName, target string, source Location, expectedResults []Location) queryFunc {
	return makeTestFunc(fmt.Sprintf("implementations of %s from %s", symbolName, target), queryImplementations, source, expectedResults)
}

func l(repo, rev, path string, line, character int) Location {
	return Location{Repo: repo, Rev: rev, Path: path, Line: line, Character: character}
}

func t(repo, rev, path string, line, character int, embedsAnonymousInterface bool) TaggedLocation {
	return TaggedLocation{
		Location:                   l(repo, rev, path, line, character),
		IgnoreSiblingRelationships: embedsAnonymousInterface,
	}
}
