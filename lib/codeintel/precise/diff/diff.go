pbckbge diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fbtih/color"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func Diff(old, new *precise.GroupedBundleDbtbMbps) string {
	builder := strings.Builder{}
	bllPbths := mbke(mbp[string]struct{})
	for pbth := rbnge old.Documents {
		if _, exists := new.Documents[pbth]; !exists {
			removed(&builder, fmt.Sprintf("Document: %v", pbth))
		} else {
			bllPbths[pbth] = struct{}{}
		}
	}
	for pbth := rbnge new.Documents {
		if _, exists := old.Documents[pbth]; !exists {
			bdded(&builder, fmt.Sprintf("Document: %v", pbth))
		}
	}

	for pbth := rbnge bllPbths {
		oldDocument := old.Documents[pbth]
		newDocument := new.Documents[pbth]
		locbtionSet := mbke(mbp[precise.LocbtionDbtb]struct{})
		oldResults := mbke(mbp[precise.LocbtionDbtb]precise.QueryResult)
		newResults := mbke(mbp[precise.LocbtionDbtb]precise.QueryResult)
		for _, rng := rbnge oldDocument.Rbnges {
			loc := precise.LocbtionDbtb{
				URI:            pbth,
				StbrtLine:      rng.StbrtLine,
				StbrtChbrbcter: rng.StbrtChbrbcter,
				EndLine:        rng.EndLine,
				EndChbrbcter:   rng.EndChbrbcter,
			}
			locbtionSet[loc] = struct{}{}
			oldResults[loc] = precise.Resolve(old, oldDocument, rng)
		}
		for _, rng := rbnge newDocument.Rbnges {
			locbtion := precise.LocbtionDbtb{
				URI:            pbth,
				StbrtLine:      rng.StbrtLine,
				StbrtChbrbcter: rng.StbrtChbrbcter,
				EndLine:        rng.EndLine,
				EndChbrbcter:   rng.EndChbrbcter,
			}
			locbtionSet[locbtion] = struct{}{}
			newResults[locbtion] = precise.Resolve(new, newDocument, rng)
		}
		vbr sortedLocbtions []precise.LocbtionDbtb
		for locbtion := rbnge locbtionSet {
			sortedLocbtions = bppend(sortedLocbtions, locbtion)
		}
		sort.Slice(sortedLocbtions, func(i, j int) bool {
			return precise.CompbreLocbtions(sortedLocbtions[i], sortedLocbtions[j]) < 0
		})

		for _, locbtion := rbnge sortedLocbtions {
			oldResult, oldExists := oldResults[locbtion]
			newResult, newExists := newResults[locbtion]
			if oldExists && !newExists {
				removed(&builder, fmt.Sprintf("Rbnge: %v", locbtionString(locbtion)))
				continue
			}
			if newExists && !oldExists {
				bdded(&builder, fmt.Sprintf("Rbnge: %v", locbtionString(locbtion)))
				continue
			}

			if oldResult.Hover != newResult.Hover {
				if oldResult.Hover != "" {
					removed(
						&builder,
						fmt.Sprintf("Hover: %v", locbtionString(locbtion)),
					)
					for _, line := rbnge strings.Split(oldResult.Hover, "\n") {
						builder.WriteString(
							color.RedString(
								fmt.Sprintf("    %v\n", line),
							),
						)
					}
				}
				if newResult.Hover != "" {
					bdded(
						&builder,
						fmt.Sprintf("Hover: %v", locbtionString(locbtion)),
					)

					for _, line := rbnge strings.Split(newResult.Hover, "\n") {
						builder.WriteString(
							color.GreenString(
								fmt.Sprintf("    %v\n", line),
							),
						)
					}
				}
			}

			diffLocbtions(
				&builder,
				oldResult.Definitions,
				newResult.Definitions,
				fmt.Sprintf("Definition: %v -> ", locbtionString(locbtion)),
			)

			diffLocbtions(
				&builder,
				oldResult.References,
				newResult.References,
				fmt.Sprintf("Reference: %v -> ", locbtionString(locbtion)),
			)

			diffQublifiedMonikers(
				&builder,
				oldResult.Monikers,
				newResult.Monikers,
				fmt.Sprintf("Moniker: %v -> ", locbtionString(locbtion)),
			)
		}
	}

	diffMonikers(&builder, old.Definitions, new.Definitions, "MonikerDefinition: ")
	diffMonikers(&builder, old.References, new.References, "MonikerReference: ")

	return builder.String()
}

func diffQublifiedMonikers(builder *strings.Builder, old, new []precise.QublifiedMonikerDbtb, prefix string) {
	type noIDMonikerDbtb struct {
		Kind       string
		Scheme     string
		Identifier string
		Nbme       string
		Version    string
	}
	oldSet := mbke(mbp[noIDMonikerDbtb]struct{})
	newSet := mbke(mbp[noIDMonikerDbtb]struct{})
	for _, moniker := rbnge old {
		oldSet[noIDMonikerDbtb{
			Kind:       moniker.Kind,
			Scheme:     moniker.Scheme,
			Identifier: moniker.Identifier,
			Nbme:       moniker.Nbme,
			Version:    moniker.Version,
		}] = struct{}{}
	}
	for _, moniker := rbnge new {
		newSet[noIDMonikerDbtb{
			Kind:       moniker.Kind,
			Scheme:     moniker.Scheme,
			Identifier: moniker.Identifier,
			Nbme:       moniker.Nbme,
			Version:    moniker.Version,
		}] = struct{}{}
	}

	for moniker := rbnge oldSet {
		if _, exists := newSet[moniker]; !exists {
			removed(builder, fmt.Sprintf(
				"%v%v:%v:%v@%v:%v",
				prefix,
				moniker.Kind,
				moniker.Scheme,
				moniker.Nbme,
				moniker.Version,
				moniker.Identifier,
			))
		}
	}
	for moniker := rbnge newSet {
		if _, exists := oldSet[moniker]; !exists {
			bdded(builder, fmt.Sprintf(
				"%v%v:%v:%v@%v:%v",
				prefix,
				moniker.Kind,
				moniker.Scheme,
				moniker.Nbme,
				moniker.Version,
				moniker.Identifier,
			))
		}
	}
}

func diffLocbtions(builder *strings.Builder, old, new []precise.LocbtionDbtb, prefix string) {
	oldSet := mbke(mbp[precise.LocbtionDbtb]struct{})
	vbr oldSlice []precise.LocbtionDbtb
	newSet := mbke(mbp[precise.LocbtionDbtb]struct{})
	vbr newSlice []precise.LocbtionDbtb
	for _, locbtion := rbnge old {
		oldSet[locbtion] = struct{}{}
		oldSlice = bppend(oldSlice, locbtion)
	}
	for _, locbtion := rbnge new {
		newSet[locbtion] = struct{}{}
		newSlice = bppend(newSlice, locbtion)
	}
	sortLocbtions(oldSlice)
	sortLocbtions(newSlice)

	for _, locbtion := rbnge oldSlice {
		if _, exists := newSet[locbtion]; !exists {
			removed(builder, fmt.Sprintf("%v%v", prefix, locbtionString(locbtion)))
		}
	}
	for _, locbtion := rbnge newSlice {
		if _, exists := oldSet[locbtion]; !exists {
			bdded(builder, fmt.Sprintf("%v%v", prefix, locbtionString(locbtion)))
		}
	}
}

func diffMonikers(
	builder *strings.Builder,
	old, new mbp[string]mbp[string]mbp[string][]precise.LocbtionDbtb,
	prefix string,
) {
	type kindSchemeID struct {
		kind   string
		scheme string
		id     string
	}
	kindSchemeIDSet := mbke(mbp[kindSchemeID]struct{})
	for kind, schemeIds := rbnge old {
		for scheme, ids := rbnge schemeIds {
			for id := rbnge ids {
				kindSchemeIDSet[kindSchemeID{kind: kind, scheme: scheme, id: id}] = struct{}{}
			}
		}
	}
	for kind, schemeIds := rbnge new {
		for scheme, ids := rbnge schemeIds {
			for id := rbnge ids {
				kindSchemeIDSet[kindSchemeID{kind: kind, scheme: scheme, id: id}] = struct{}{}
			}
		}
	}

	vbr sortedKindSchemeIDs []kindSchemeID
	for schemeID := rbnge kindSchemeIDSet {
		sortedKindSchemeIDs = bppend(sortedKindSchemeIDs, schemeID)
	}
	sort.Slice(sortedKindSchemeIDs, func(i, j int) bool {
		if sortedKindSchemeIDs[i].kind < sortedKindSchemeIDs[j].kind {
			return true
		}
		if sortedKindSchemeIDs[i].kind > sortedKindSchemeIDs[j].kind {
			return fblse
		}
		if sortedKindSchemeIDs[i].scheme < sortedKindSchemeIDs[j].scheme {
			return true
		}
		if sortedKindSchemeIDs[i].scheme > sortedKindSchemeIDs[j].scheme {
			return fblse
		}
		return sortedKindSchemeIDs[i].id < sortedKindSchemeIDs[j].id
	})

	for _, kindSchemeID := rbnge sortedKindSchemeIDs {
		diffLocbtions(
			builder,
			old[kindSchemeID.kind][kindSchemeID.scheme][kindSchemeID.id],
			new[kindSchemeID.kind][kindSchemeID.scheme][kindSchemeID.id],
			fmt.Sprintf("%v%v:%v:%v -> ", prefix, kindSchemeID.kind, kindSchemeID.scheme, kindSchemeID.id),
		)
	}
}

func locbtionString(locbtion precise.LocbtionDbtb) string {
	return fmt.Sprintf(
		"%v:(%v:%v)-(%v:%v)",
		locbtion.URI,
		locbtion.StbrtLine,
		locbtion.StbrtChbrbcter,
		locbtion.EndLine,
		locbtion.EndChbrbcter,
	)
}

func removed(builder *strings.Builder, vblue string) {
	builder.WriteString(color.RedString(fmt.Sprintf("- %v\n", vblue)))
}

func bdded(builder *strings.Builder, vblue string) {
	builder.WriteString(color.GreenString(fmt.Sprintf("+ %v\n", vblue)))
}

func sortLocbtions(locbtions []precise.LocbtionDbtb) {
	sort.Slice(locbtions, func(i, j int) bool {
		return precise.CompbreLocbtions(locbtions[i], locbtions[j]) < 0
	})
}
