import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, ViewPlugin, ViewUpdate } from '@codemirror/view'

export type HighlightRange = [number, number, string]

export const highlightRanges = Facet.define<HighlightRange[], HighlightRange[]>({
    combine(ranges) {
        return ranges.flat()
    },
    enables: facet =>
        ViewPlugin.fromClass(
            class {
                decorationCache: Record<string, Decoration> = {}
                decorations: DecorationSet = Decoration.none

                constructor(view: EditorView) {
                    this.decorations = this.computeDecorations(view)
                }

                update(update: ViewUpdate) {
                    if (update.docChanged) {
                        this.decorationCache = {}
                    }

                    if (update.viewportChanged) {
                        this.decorations = this.computeDecorations(update.view)
                    }
                }

                computeDecorations(view: EditorView): DecorationSet {
                    console.log('update')
                    const { from, to } = view.viewport
                    const ranges = view.state.facet(facet)
                    const rangeIndex = rangeIndexOf(ranges, from)

                    if (rangeIndex === -1) {
                        return Decoration.none
                    }
                    const builder = new RangeSetBuilder<Decoration>()

                    for (let index = rangeIndex; index < ranges.length && ranges[index][0] <= to; index++) {
                        const [start, end, cls] = ranges[index]
                        builder.add(
                            start,
                            end,
                            this.decorationCache[cls] || (this.decorationCache[cls] = Decoration.mark({ class: cls }))
                        )
                    }

                    return builder.finish()
                }
            },
            { decorations: plugin => plugin.decorations }
        ),
})

/**
 * Performs a binary search to find the left most range whose end is start or
 * the right most element whose end is < start.
 */
function rangeIndexOf(ranges: [number, number, unknown][], start: number): number | -1 {
    let low = 0
    let high = ranges.length

    while (low < high) {
        const middle = Math.floor((low + high) / 2)
        // It uses the end of the range for comparison because we want all
        //kdecorations that are applicable at a certain position.
        if (ranges[middle][1] < start) {
            low = middle + 1
        } else {
            high = middle
        }
    }

    return ranges[low] === undefined ? -1 : low
}
