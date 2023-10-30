import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration } from '@codemirror/view'
import { from } from 'rxjs'

import { getDocumentHighlights } from './api'
import { codeIntelDecorations } from './decorations'
import { UpdateableValue, createLoaderExtension } from './utils'

interface Range {
    from: number
    to: number
}

const documentHighlightDeco = Decoration.mark({ class: 'sourcegraph-document-highlight' })

class Highlights implements UpdateableValue<Range[], Highlights> {
    constructor(public range: Range, public highlights: Range[] | null) {}

    update(highlights: Range[]) {
        return new Highlights(this.range, highlights)
    }

    get isPending() {
        return this.highlights === null
    }

    get key() {
        return this.range
    }
}

/**
 * Facet to register for which ranges to show document highlights.
 */
export const showDocumentHighlights: Facet<Range> = Facet.define<Range>({
    enables: self => [
        createLoaderExtension({
            input(state) {
                return state.facet(self)
            },
            create(range) {
                return new Highlights(range, null)
            },
            load(highlights, state) {
                return from(getDocumentHighlights(state, highlights.range.from))
            },
            provide: self => [
                codeIntelDecorations.computeN([self], state =>
                    state.field(self).map(({ highlights }) => {
                        if (highlights) {
                            const builder = new RangeSetBuilder<Decoration>()
                            for (const highlight of highlights) {
                                builder.add(highlight.from, highlight.to, documentHighlightDeco)
                            }
                            return builder.finish()
                        }
                        return Decoration.none
                    })
                ),
            ],
        }),
    ],
})
