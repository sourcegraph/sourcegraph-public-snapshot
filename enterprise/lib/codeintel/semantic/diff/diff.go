package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/semantic"
)

func Diff(old, new *semantic.GroupedBundleDataMaps) string {
	builder := strings.Builder{}
	allPaths := make(map[string]struct{})
	for path := range old.Documents {
		if _, exists := new.Documents[path]; !exists {
			removed(&builder, fmt.Sprintf("Document: %v", path))
		} else {
			allPaths[path] = struct{}{}
		}
	}
	for path := range new.Documents {
		if _, exists := old.Documents[path]; !exists {
			added(&builder, fmt.Sprintf("Document: %v", path))
		}
	}

	for path := range allPaths {
		oldDocument := old.Documents[path]
		newDocument := new.Documents[path]
		locationSet := make(map[semantic.LocationData]struct{})
		oldResults := make(map[semantic.LocationData]semantic.QueryResult)
		newResults := make(map[semantic.LocationData]semantic.QueryResult)
		for _, rng := range oldDocument.Ranges {
			loc := semantic.LocationData{
				URI:            path,
				StartLine:      rng.StartLine,
				StartCharacter: rng.StartCharacter,
				EndLine:        rng.EndLine,
				EndCharacter:   rng.EndCharacter,
			}
			locationSet[loc] = struct{}{}
			oldResults[loc] = semantic.Resolve(old, oldDocument, rng)
		}
		for _, rng := range newDocument.Ranges {
			location := semantic.LocationData{
				URI:            path,
				StartLine:      rng.StartLine,
				StartCharacter: rng.StartCharacter,
				EndLine:        rng.EndLine,
				EndCharacter:   rng.EndCharacter,
			}
			locationSet[location] = struct{}{}
			newResults[location] = semantic.Resolve(new, newDocument, rng)
		}
		var sortedLocations []semantic.LocationData
		for location := range locationSet {
			sortedLocations = append(sortedLocations, location)
		}
		sort.Slice(sortedLocations, func(i, j int) bool {
			return semantic.CompareLocations(sortedLocations[i], sortedLocations[j]) < 0
		})

		for _, location := range sortedLocations {
			oldResult, oldExists := oldResults[location]
			newResult, newExists := newResults[location]
			if oldExists && !newExists {
				removed(&builder, fmt.Sprintf("Range: %v", locationString(location)))
				continue
			}
			if newExists && !oldExists {
				added(&builder, fmt.Sprintf("Range: %v", locationString(location)))
				continue
			}

			if oldResult.Hover != newResult.Hover {
				if oldResult.Hover != "" {
					removed(
						&builder,
						fmt.Sprintf("Hover: %v", locationString(location)),
					)
					for _, line := range strings.Split(oldResult.Hover, "\n") {
						builder.WriteString(
							color.RedString(
								fmt.Sprintf("    %v\n", line),
							),
						)
					}
				}
				if newResult.Hover != "" {
					added(
						&builder,
						fmt.Sprintf("Hover: %v", locationString(location)),
					)

					for _, line := range strings.Split(newResult.Hover, "\n") {
						builder.WriteString(
							color.GreenString(
								fmt.Sprintf("    %v\n", line),
							),
						)
					}
				}
			}

			diffLocations(
				&builder,
				oldResult.Definitions,
				newResult.Definitions,
				fmt.Sprintf("Definition: %v -> ", locationString(location)),
			)

			diffLocations(
				&builder,
				oldResult.References,
				newResult.References,
				fmt.Sprintf("Reference: %v -> ", locationString(location)),
			)

			diffQualifiedMonikers(
				&builder,
				oldResult.Monikers,
				newResult.Monikers,
				fmt.Sprintf("Moniker: %v -> ", locationString(location)),
			)
		}
	}

	diffMonikers(&builder, old.Definitions, new.Definitions, "MonikerDefinition: ")
	diffMonikers(&builder, old.References, new.References, "MonikerReference: ")

	return builder.String()
}

func diffQualifiedMonikers(builder *strings.Builder, old, new []semantic.QualifiedMonikerData, prefix string) {
	type noIDMonikerData struct {
		Kind       string
		Scheme     string
		Identifier string
		Name       string
		Version    string
	}
	oldSet := make(map[noIDMonikerData]struct{})
	newSet := make(map[noIDMonikerData]struct{})
	for _, moniker := range old {
		oldSet[noIDMonikerData{
			Kind:       moniker.Kind,
			Scheme:     moniker.Scheme,
			Identifier: moniker.Identifier,
			Name:       moniker.Name,
			Version:    moniker.Version,
		}] = struct{}{}
	}
	for _, moniker := range new {
		newSet[noIDMonikerData{
			Kind:       moniker.Kind,
			Scheme:     moniker.Scheme,
			Identifier: moniker.Identifier,
			Name:       moniker.Name,
			Version:    moniker.Version,
		}] = struct{}{}
	}

	for moniker := range oldSet {
		if _, exists := newSet[moniker]; !exists {
			removed(builder, fmt.Sprintf(
				"%v%v:%v:%v@%v:%v",
				prefix,
				moniker.Kind,
				moniker.Scheme,
				moniker.Name,
				moniker.Version,
				moniker.Identifier,
			))
		}
	}
	for moniker := range newSet {
		if _, exists := oldSet[moniker]; !exists {
			added(builder, fmt.Sprintf(
				"%v%v:%v:%v@%v:%v",
				prefix,
				moniker.Kind,
				moniker.Scheme,
				moniker.Name,
				moniker.Version,
				moniker.Identifier,
			))
		}
	}
}

func diffLocations(builder *strings.Builder, old, new []semantic.LocationData, prefix string) {
	oldSet := make(map[semantic.LocationData]struct{})
	var oldSlice []semantic.LocationData
	newSet := make(map[semantic.LocationData]struct{})
	var newSlice []semantic.LocationData
	for _, location := range old {
		oldSet[location] = struct{}{}
		oldSlice = append(oldSlice, location)
	}
	for _, location := range new {
		newSet[location] = struct{}{}
		newSlice = append(newSlice, location)
	}
	sortLocations(oldSlice)
	sortLocations(newSlice)

	for _, location := range oldSlice {
		if _, exists := newSet[location]; !exists {
			removed(builder, fmt.Sprintf("%v%v", prefix, locationString(location)))
		}
	}
	for _, location := range newSlice {
		if _, exists := oldSet[location]; !exists {
			added(builder, fmt.Sprintf("%v%v", prefix, locationString(location)))
		}
	}
}

func diffMonikers(
	builder *strings.Builder,
	old, new map[string]map[string][]semantic.LocationData,
	prefix string,
) {
	type schemeID struct {
		scheme string
		id     string
	}
	schemeIDSet := make(map[schemeID]struct{})
	for scheme, ids := range old {
		for id := range ids {
			schemeIDSet[schemeID{scheme: scheme, id: id}] = struct{}{}
		}
	}
	for scheme, ids := range new {
		for id := range ids {
			schemeIDSet[schemeID{scheme: scheme, id: id}] = struct{}{}
		}
	}

	var sortedSchemeIDs []schemeID
	for schemeID := range schemeIDSet {
		sortedSchemeIDs = append(sortedSchemeIDs, schemeID)
	}
	sort.Slice(sortedSchemeIDs, func(i, j int) bool {
		if sortedSchemeIDs[i].scheme < sortedSchemeIDs[j].scheme {
			return true
		}
		if sortedSchemeIDs[i].scheme > sortedSchemeIDs[j].scheme {
			return false
		}
		return sortedSchemeIDs[i].id < sortedSchemeIDs[j].id
	})

	for _, schemeID := range sortedSchemeIDs {
		diffLocations(
			builder,
			old[schemeID.scheme][schemeID.id],
			new[schemeID.scheme][schemeID.id],
			fmt.Sprintf("%v%v:%v -> ", prefix, schemeID.scheme, schemeID.id),
		)
	}
}

func locationString(location semantic.LocationData) string {
	return fmt.Sprintf(
		"%v:(%v:%v)-(%v:%v)",
		location.URI,
		location.StartLine,
		location.StartCharacter,
		location.EndLine,
		location.EndCharacter,
	)
}

func removed(builder *strings.Builder, value string) {
	builder.WriteString(color.RedString(fmt.Sprintf("- %v\n", value)))
}

func added(builder *strings.Builder, value string) {
	builder.WriteString(color.GreenString(fmt.Sprintf("+ %v\n", value)))
}

func sortLocations(locations []semantic.LocationData) {
	sort.Slice(locations, func(i, j int) bool {
		return semantic.CompareLocations(locations[i], locations[j]) < 0
	})
}
