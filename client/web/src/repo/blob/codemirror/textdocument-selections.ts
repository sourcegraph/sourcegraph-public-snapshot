import { Facet, Line, RangeSet, StateField } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'

import { Occurrence, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { shouldScrollIntoView } from './linenumbers'
import { cmSelectionToRange, occurrenceAtPosition, rangeToSelection } from './positions'
import { hoverAtOccurrence } from './textdocument-hover'
import { blobInfoFacet, historyFacet, selectionsFacet, uriFacet } from './textdocument-facets'
import { goToDefinitionAtOccurrence } from './textdocument-definition'

export const tokenSelectionTheme = EditorView.theme({
    '.cm-token-selection-clickable': {
        cursor: 'pointer',
    },
    '.cm-token-selection-definition-ready': {
        textDecoration: 'underline',
    },
})

const ariaCurrent = Decoration.mark({
    attributes: { 'aria-current': 'true' },
})
const selectedOccurrenceField = StateField.define<Occurrence | undefined>({
    create: () => undefined,
    update(value, transaction) {
        const selection = transaction.selection ?? transaction.state.selection
        const position = cmSelectionToRange(transaction.state, selection.main)
        const atPosition = occurrenceAtPosition(transaction.state, position.start)
        if (atPosition) {
            return atPosition.occurrence
        }
        return undefined
    },
})
export const selectedOccurrence = Facet.define<unknown, unknown>({
    combine: sources => sources[0],
    enables: () => [
        EditorView.decorations.compute([selectedOccurrenceField], state => {
            const occ = state.field(selectedOccurrenceField)
            if (occ) {
                const range = rangeToSelection(state, occ.range)
                return RangeSet.of([ariaCurrent.range(range.from, range.to)])
            }
            return RangeSet.empty
        }),
        selectedOccurrenceField,
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

export const selectOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    const blobInfo = view.state.facet(blobInfoFacet)
    hoverAtOccurrence(view, occurrence).then(
        () => {},
        () => {}
    )
    goToDefinitionAtOccurrence(view, occurrence.range.start, occurrence).then(
        () => {},
        () => {}
    )
    const url = toPrettyBlobURL({ ...blobInfo, range: occurrence.range.asOneBased() })
    view.state.facet(historyFacet).replace(url)
    selectRange(view, occurrence.range)
}

export const selectRange = (view: EditorView, range: Range): void => {
    const selection = rangeToSelection(view.state, range)
    const uri = view.state.facet(uriFacet)
    const selections = view.state.facet(selectionsFacet)
    if (selections) {
        selections.set(uri, range)
    }
    view.dispatch({ selection })

    const lineAbove = view.state.doc.line(Math.min(view.state.doc.lines, range.start.line + 3))
    if (scrollLineIntoView(view, lineAbove)) {
        return
    }
    const lineBelow = view.state.doc.line(Math.max(1, range.end.line - 1))
    scrollLineIntoView(view, lineBelow)
}
