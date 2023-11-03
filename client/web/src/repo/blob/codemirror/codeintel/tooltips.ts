import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Tooltip, showTooltip as showCodeMirrorTooltip, Decoration } from '@codemirror/view'
import { Observable, isObservable } from 'rxjs'

import { codeIntelDecorations } from './decorations'
import { UpdateableValue, createLoaderExtension } from './utils'

export type TooltipSource = Observable<Tooltip | null> | Tooltip | null

enum Status {
    PENDING,
    DONE,
}

/**
 * This facet deduplicates tooltips shown at the same position. Only the first registered
 * one for a given position will be shown. This is done here so that different sources
 * that provide tooltips don't need to depend on each other and check whether or not to
 * show a tooltip. The order is determined by the order of extensions that provide values
 * for the {@link showTooltip} facet.
 * This facet enables token styling via {@link codeIntelDecorations} and passes the final
 * list of tooltips to CodeMirror's own `showTooltip` facet.
 */
const uniqueTooltips = Facet.define<Tooltip | null, Tooltip[]>({
    combine(values) {
        const seen = new Set<number>()
        const result = []
        for (const value of values) {
            if (value && !seen.has(value.pos)) {
                seen.add(value.pos)
                result.push(value)
            }
        }
        return result.sort((a, b) => a.pos - b.pos)
    },
    compare(a, b) {
        return a.length === b.length && a.every((value, i) => value === b[i])
    },
    enables: self => [
        // Highlight tokens with tooltips
        codeIntelDecorations.compute([self], state => {
            let decorations = new RangeSetBuilder<Decoration>()
            const tooltips = state.facet(self)
            for (const { pos, end } of tooltips) {
                if (end && pos !== end) {
                    decorations.add(pos, end, Decoration.mark({ class: `selection-highlight` }))
                }
            }

            return decorations.finish()
        }),

        // Show tooltips
        showCodeMirrorTooltip.computeN([self], state => state.facet(self)),
    ],
})

/**
 * Class for keeping track of the currently shown tooltip at the specified position.
 * The class is designed to allow showing multiple tooltips over time, which allows
 * for features like a temporary loading tooltip.
 */
class TooltipLoadingState implements UpdateableValue<Tooltip | null, TooltipLoadingState> {
    public visible: boolean

    constructor(public source: TooltipSource, public status: Status, public tooltip: Tooltip | null = null) {
        this.visible = !!tooltip
    }

    update(tooltip: Tooltip | null) {
        return new TooltipLoadingState(this.source, Status.DONE, tooltip)
    }

    get key() {
        return this.source
    }

    get isPending() {
        return this.status === Status.PENDING
    }
}

/**
 * Facet for registring where to show tooltips.
 */
export const showTooltip: Facet<TooltipSource> = Facet.define<TooltipSource>({
    enables: self => [
        createLoaderExtension({
            input(state) {
                return state.facet(self)
            },
            create(source) {
                return source && isObservable(source)
                    ? new TooltipLoadingState(source, Status.PENDING)
                    : new TooltipLoadingState(source, Status.DONE, source)
            },
            load(value) {
                return value.source as Observable<Tooltip | null>
            },
            provide: self => [
                uniqueTooltips.computeN([self], state => state.field(self).map(tooltip => tooltip.tooltip)),
            ],
        }),
    ],
})
