import { type SelectionRange, StateField } from '@codemirror/state'
import type { EditorView } from '@codemirror/view'
import { memoize } from 'lodash'
import { type Location, createPath } from 'react-router-dom'

import { type Occurrence, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { cmSelectionToRange, occurrenceAtPosition, rangeToCmSelection } from '../occurrence-utils'
import { isSelectionInsideDocument } from '../utils'

import { setFocusedOccurrence } from './code-intel-tooltips'

export const fallbackOccurrences = StateField.define<Map<number, Occurrence>>({
    create: () => new Map(),
    update: value => value,
})

// Helper function to update editor selection from URL information.
export const syncOccurrencesWithURL = memoize(
    (location: Location, view: EditorView) => {
        const { selection } = selectionFromLocation(view, location)

        if (selection && isSelectionInsideDocument(selection, view.state.doc)) {
            const occurrence = occurrenceAtPosition(view.state, cmSelectionToRange(view.state, selection).start)

            view.dispatch({ effects: setFocusedOccurrence.of(occurrence ?? null) })
        }
    },
    location => createPath(location)
)

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
