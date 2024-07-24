package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func Diff(old, new *precise.GroupedBundleDataMaps) string {
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
		locationSet := make(map[precise.LocationData]struct{})
		oldResults := make(map[precise.LocationData]precise.QueryResult)
		newResults := make(map[precise.LocationData]precise.QueryResult)
		for _, rng := range oldDocument.Ranges {
			loc := precise.LocationData{
				DocumentPath:   path,
				StartLine:      rng.StartLine,
				StartCharacter: rng.StartCharacter,
				EndLine:        rng.EndLine,
				EndCharacter:   rng.EndCharacter,
			}
			locationSet[loc] = struct{}{}
			oldResults[loc] = precise.Resolve(old, oldDocument, rng)
		}
		for _, rng := range newDocument.Ranges {
			location := precise.LocationData{
				DocumentPath:   path,
				StartLine:      rng.StartLine,
				StartCharacter: rng.StartCharacter,
				EndLine:        rng.EndLine,
				EndCharacter:   rng.EndCharacter,
			}
			locationSet[location] = struct{}{}
			newResults[location] = precise.Resolve(new, newDocument, rng)
		}
		var sortedLocations []precise.LocationData
		for location := range locationSet {
			sortedLocations = append(sortedLocations, location)
		}
		sort.Slice(sortedLocations, func(i, j int) bool {
			return precise.CompareLocations(sortedLocations[i], sortedLocations[j]) < 0
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

func diffQualifiedMonikers(builder *strings.Builder, old, new []precise.QualifiedMonikerData, prefix string) {
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

func diffLocations(builder *strings.Builder, old, new []precise.LocationData, prefix string) {
	oldSet := make(map[precise.LocationData]struct{})
	var oldSlice []precise.LocationData
	newSet := make(map[precise.LocationData]struct{})
	var newSlice []precise.LocationData
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
	old, new map[string]map[string]map[string][]precise.LocationData,
	prefix string,
) {
	type kindSchemeID struct {
		kind   string
		scheme string
		id     string
	}
	kindSchemeIDSet := make(map[kindSchemeID]struct{})
	for kind, schemeIds := range old {
		for scheme, ids := range schemeIds {
			for id := range ids {
				kindSchemeIDSet[kindSchemeID{kind: kind, scheme: scheme, id: id}] = struct{}{}
			}
		}
	}
	for kind, schemeIds := range new {
		for scheme, ids := range schemeIds {
			for id := range ids {
				kindSchemeIDSet[kindSchemeID{kind: kind, scheme: scheme, id: id}] = struct{}{}
			}
		}
	}

	var sortedKindSchemeIDs []kindSchemeID
	for schemeID := range kindSchemeIDSet {
		sortedKindSchemeIDs = append(sortedKindSchemeIDs, schemeID)
	}
	sort.Slice(sortedKindSchemeIDs, func(i, j int) bool {
		if sortedKindSchemeIDs[i].kind < sortedKindSchemeIDs[j].kind {
			return true
		}
		if sortedKindSchemeIDs[i].kind > sortedKindSchemeIDs[j].kind {
			return false
		}
		if sortedKindSchemeIDs[i].scheme < sortedKindSchemeIDs[j].scheme {
			return true
		}
		if sortedKindSchemeIDs[i].scheme > sortedKindSchemeIDs[j].scheme {
			return false
		}
		return sortedKindSchemeIDs[i].id < sortedKindSchemeIDs[j].id
	})

	for _, kindSchemeID := range sortedKindSchemeIDs {
		diffLocations(
			builder,
			old[kindSchemeID.kind][kindSchemeID.scheme][kindSchemeID.id],
			new[kindSchemeID.kind][kindSchemeID.scheme][kindSchemeID.id],
			fmt.Sprintf("%v%v:%v:%v -> ", prefix, kindSchemeID.kind, kindSchemeID.scheme, kindSchemeID.id),
		)
	}
}

func locationString(location precise.LocationData) string {
	return fmt.Sprintf(
		"%v:(%v:%v)-(%v:%v)",
		location.DocumentPath,
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

func sortLocations(locations []precise.LocationData) {
	sort.Slice(locations, func(i, j int) bool {
		return precise.CompareLocations(locations[i], locations[j]) < 0
	})
}
