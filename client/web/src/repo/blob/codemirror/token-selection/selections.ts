import { Extension, Facet, Line, RangeSet, SelectionRange, StateField } from '@codemirror/state'
import { Decoration, EditorView, PluginValue, ViewPlugin } from '@codemirror/view'
import * as H from 'history'

import { Occurrence, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { shouldScrollIntoView } from '../linenumbers'
import { cmSelectionToRange, occurrenceAtPosition, rangeToCmSelection } from '../occurrence-utils'
import { isSelectionInsideDocument } from '../utils'

import { definitionCache, goToDefinitionAtOccurrence } from './definition'
import { showDocumentHighlightsForOccurrence } from './document-highlights'
import { hoverAtOccurrence, hoverCache, setHoveredOccurrenceEffect } from './hover'

export const tokenSelectionTheme = EditorView.theme({
    '.cm-token-selection-definition-ready': {
        textDecoration: 'underline',
    },
    '.cm-token-selection-clickable:hover': {
        cursor: 'pointer',
    },
})

const ariaCurrent = Decoration.mark({
    attributes: { 'aria-current': 'true' },
})

export const fallbackOccurrences = StateField.define<Map<number, Occurrence>>({
    create: () => new Map(),
    update: value => value,
})
const selectedOccurrenceField = StateField.define<Occurrence | undefined>({
    create: () => undefined,
    update(value, transaction) {
        const selection = transaction.selection ?? transaction.state.selection
        const position = cmSelectionToRange(transaction.state, selection.main)
        const occurrence = occurrenceAtPosition(transaction.state, position.start)
        if (occurrence) {
            return occurrence
        }
        return undefined
    },
})
export const selectedOccurrence = Facet.define<unknown, unknown>({
    combine: sources => sources[0],
    enables: () => [
        fallbackOccurrences,
        selectedOccurrenceField,
        EditorView.decorations.compute([selectedOccurrenceField], state => {
            const occ = state.field(selectedOccurrenceField)
            if (occ) {
                const range = rangeToCmSelection(state, occ.range)
                if (range.from === range.to) {
                    return RangeSet.empty
                }
                return RangeSet.of([ariaCurrent.range(range.from, range.to)])
            }
            return RangeSet.empty
        }),
    ],
})

const scrollLineIntoView = (view: EditorView, line: Line): boolean => {
    if (shouldScrollIntoView(view, { line: line.number })) {
        view.dispatch({
            effects: EditorView.scrollIntoView(line.from, { y: 'nearest' }),
        })
        return true
    }
    return false
}

export const scrollRangeIntoView = (view: EditorView, range: Range): void => {
    const lineAbove = view.state.doc.line(Math.min(view.state.doc.lines, range.start.line + 3))
    if (scrollLineIntoView(view, lineAbove)) {
        return
    }
    const lineBelow = view.state.doc.line(Math.max(1, range.end.line - 1))
    scrollLineIntoView(view, lineBelow)
}

export const warmupOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    if (!view.state.field(hoverCache).has(occurrence)) {
        hoverAtOccurrence(view, occurrence).then(
            () => {},
            () => {}
        )
    }
    if (!view.state.field(definitionCache).has(occurrence)) {
        goToDefinitionAtOccurrence(view, occurrence).then(
            () => {},
            () => {}
        )
    }
}

export const selectOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    warmupOccurrence(view, occurrence)
    showDocumentHighlightsForOccurrence(view, occurrence)
    selectRange(view, occurrence.range)
    view.dispatch({ effects: setHoveredOccurrenceEffect.of(occurrence) })
}

export const selectRange = (view: EditorView, range: Range): void => {
    const selection = rangeToCmSelection(view.state, range)
    view.dispatch({ selection })
    scrollRangeIntoView(view, range)
}

// View plugin that listens to history location changes and updates editor
// selection accordingly.
export const syncSelectionWithURL: Extension = ViewPlugin.fromClass(
    class implements PluginValue {
        private onDestroy: H.UnregisterCallback
        constructor(public view: EditorView) {
            const history = view.state.facet(blobPropsFacet).history
            this.onDestroy = history.listen(location => this.onLocation(location))
        }
        public onLocation(location: H.Location): void {
            const { selection } = selectionFromLocation(this.view, location)
            if (selection && isSelectionInsideDocument(selection, this.view.state.doc)) {
                this.view.dispatch({ selection })
            }
        }
        public destroy(): void {
            this.onDestroy()
        }
    }
)

export function selectionFromLocation(
    view: EditorView,
    location: H.Location
): { range?: Range; selection?: SelectionRange } {
    const { line, character, endCharacter } = parseQueryAndHash(location.search, location.hash)
    if (line && character && endCharacter) {
        const range = Range.fromNumbers(line, character, line, endCharacter).withDecrementedValues()
        const selection = rangeToCmSelection(view.state, range)
        return { range, selection }
    }
    return {}
}
