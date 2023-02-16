import { Extension, SelectionRange, StateField } from '@codemirror/state'
import { EditorView, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { memoize } from 'lodash'
import { Location, createPath } from 'react-router-dom-v5-compat'

import { Occurrence, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { cmSelectionToRange, occurrenceAtPosition, rangeToCmSelection } from '../occurrence-utils'
import { isSelectionInsideDocument } from '../utils'

import { setFocusedOccurrence } from './code-intel-tooltips'

export const fallbackOccurrences = StateField.define<Map<number, Occurrence>>({
    create: () => new Map(),
    update: value => value,
})

// View plugin that listens to location changes and updates editor selection accordingly.
export const syncOccurrenceWithURL: Extension = ViewPlugin.define(view => ({
    update(update: ViewUpdate): void {
        const { location } = update.state.facet(blobPropsFacet)

        // Update occurences only on location change.
        this.updateFocusedOccurences(createPath(location), location)
    },
    // The first argument of the memoized function is used as a cache key.
    updateFocusedOccurences: memoize((pathCacheKey: string, location: Location) => {
        const { selection } = selectionFromLocation(view, location)

        if (selection && isSelectionInsideDocument(selection, view.state.doc)) {
            const occurrence = occurrenceAtPosition(view.state, cmSelectionToRange(view.state, selection).start)

            window.requestAnimationFrame(() => view.dispatch({ effects: setFocusedOccurrence.of(occurrence ?? null) }))
        }
    }),
}))

export function selectionFromLocation(
    view: EditorView,
    location: Location
): { range?: Range; selection?: SelectionRange } {
    const { line, character, endCharacter } = parseQueryAndHash(location.search, location.hash)
    if (line && character && endCharacter) {
        const range = Range.fromNumbers(line, character, line, endCharacter).withDecrementedValues()
        const selection = rangeToCmSelection(view.state, range)
        return { range, selection }
    }
    return {}
}
