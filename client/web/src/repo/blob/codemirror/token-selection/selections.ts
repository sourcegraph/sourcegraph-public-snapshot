import { Extension, SelectionRange, StateField } from '@codemirror/state'
import { EditorView, PluginValue, ViewPlugin } from '@codemirror/view'
import * as H from 'history'

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

// View plugin that listens to history location changes and updates editor
// selection accordingly.
export const syncOccurrenceWithURL: Extension = ViewPlugin.fromClass(
    class implements PluginValue {
        private onDestroy: H.UnregisterCallback
        constructor(public view: EditorView) {
            // TODO(valery): RR6
            const history = view.state.facet(blobPropsFacet).history
            this.onDestroy = history.listen(location => this.onLocation(location))
        }
        public onLocation(location: H.Location): void {
            const { selection } = selectionFromLocation(this.view, location)
            if (selection && isSelectionInsideDocument(selection, this.view.state.doc)) {
                const occurrence = occurrenceAtPosition(
                    this.view.state,
                    cmSelectionToRange(this.view.state, selection).start
                )

                this.view.dispatch({ effects: setFocusedOccurrence.of(occurrence ?? null) })
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
