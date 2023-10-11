import { type SelectionRange } from '@codemirror/state'
import type { EditorView } from '@codemirror/view'
import { type Location } from 'react-router-dom'

import { Range } from '@sourcegraph/shared/src/codeintel/scip'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { rangeToCmSelection } from '../occurrence-utils'

import { setSelection } from './token-selection'

// Helper function to update editor selection from URL information.
export const syncOccurrencesWithURL = (location: Location, view: EditorView) => {
    const { selection } = selectionFromLocation(view, location)
    if (selection) {
        view.dispatch(setSelection(selection.from))
    }
}

export function selectionFromLocation(
    view: EditorView,
    location: Location
): { range?: Range; selection?: SelectionRange } {
    const { line, character, endCharacter } = parseQueryAndHash(location.search, location.hash)
    if (line && character && endCharacter) {
        const range = Range.fromNumbers(line, character, line, endCharacter).withDecrementedValues()
        const selection = rangeToCmSelection(view.state.doc, range)
        return { range, selection }
    }
    return {}
}
