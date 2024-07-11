import { Facet, StateField, Transaction, countColumn } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, WidgetType } from '@codemirror/view'

import { Occurrence, SymbolRole } from '@sourcegraph/shared/src/codeintel/scip'

class DebugOccurrencesWidget extends WidgetType {
    constructor(private occurrence: Occurrence) {
        super()
    }

    public toDOM(view: EditorView): HTMLElement {
        const div = document.createElement('div')
        div.style.color = 'grey'
        // Calculate the visual offset of the range. This is not the same as start character
        // because tabs take up multiple spaces visually.
        const content = view.state.doc.line(this.occurrence.range.start.line + 1).text
        const spaceCount = countColumn(
            content,
            view.state.tabSize,
            Math.min(this.occurrence.range.start.character, content.length)
        )
        const spaces = ' '.repeat(spaceCount)
        const arrows = '^'.repeat(this.occurrence.range.end.character - this.occurrence.range.start.character)
        const type = (this.occurrence.symbolRoles ?? 0) & SymbolRole.Definition ? 'definition' : 'reference'
        const symbolName = this.occurrence.symbol ?? ''
        div.innerText = `${spaces}${arrows} ${type} ${symbolName}`
        return div
    }
}

export const debugOccurrencesDecorations = StateField.define<DecorationSet>({
    create() {
        return Decoration.none
    },
    update(oldDecorations: DecorationSet, tr: Transaction): DecorationSet {
        const oldOccurrences = tr.startState.facet(debugOccurrences)
        const newOccurrences = tr.state.facet(debugOccurrences)
        if (Object.is(newOccurrences, oldOccurrences)) {
            return oldDecorations
        }
        return Decoration.set(
            newOccurrences.map(occ => {
                // Place the block at the end of the line so it doesn't split the line
                const rangeStart = tr.state.doc.line(occ.range.start.line + 1).to + 1
                return Decoration.widget({
                    widget: new DebugOccurrencesWidget(occ),
                    block: true,
                }).range(rangeStart, rangeStart)
            }),
            true // TODO(camdencheek): we can usually guarantee this is sorted, but I'd rather use a type for that
        )
    },
    provide: f => EditorView.decorations.from(f),
})

export const debugOccurrences = Facet.define<Occurrence[], Occurrence[]>({
    combine(input: readonly Occurrence[][]): Occurrence[] {
        return input[0] ?? []
    },
    enables: () => debugOccurrencesDecorations,
})
