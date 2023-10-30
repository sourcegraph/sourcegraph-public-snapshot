import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Tooltip, showTooltip as showCodeMirrorTooltip, EditorView, Decoration } from '@codemirror/view'
import { Observable, isObservable } from 'rxjs'

import { CodeIntelTooltipPosition } from './api'
import { codeIntelDecorations } from './decorations'
import { UpdateableValue, createLoaderExtension } from './utils'

type TooltipWithEnd = Tooltip & { end: number }
export type TooltipSource = TooltipWithEnd | Observable<TooltipWithEnd | null>

export interface CodeIntelTooltip {
    range: CodeIntelTooltipPosition
    source: TooltipSource
    key: string
}

enum Status {
    PENDING,
    DONE,
}

const uniqueTooltips = Facet.define<TooltipWithEnd | null, TooltipWithEnd[]>({
    combine(values) {
        const seen = new Set<number>()
        const result = []
        for (const value of values) {
            if (value && !seen.has(value.pos)) {
                seen.add(value.pos)
                result.push(value)
            }
        }
        return result.sort((a, b) => a.pos - b.end)
    },
    enables: self => [
        EditorView.theme({
            '.cm-tooltip.sg-code-intel-hovercard': {
                border: 'unset',
            },
        }),

        // Highlight tokens with tooltips
        codeIntelDecorations.compute([self], state => {
            let decorations = new RangeSetBuilder<Decoration>()
            const tooltips = state.facet(self)
            for (const { pos, end } of tooltips) {
                decorations.add(pos, end, Decoration.mark({ class: `selection-highlight` }))
            }

            return decorations.finish()
        }),

        // Show tooltips
        showCodeMirrorTooltip.computeN([self], state => state.facet(self)),
    ],
})

class TooltipLoadingState implements UpdateableValue<TooltipWithEnd | null, TooltipLoadingState> {
    public visible: boolean

    constructor(
        public codeIntelTooltip: CodeIntelTooltip,
        public status: Status,
        public tooltip: TooltipWithEnd | null = null
    ) {
        this.visible = !!tooltip
    }

    update(tooltip: TooltipWithEnd | null) {
        return new TooltipLoadingState(this.codeIntelTooltip, Status.DONE, tooltip)
    }

    get key() {
        return this.codeIntelTooltip.source
    }

    get isPending() {
        return this.status === Status.PENDING
    }
}

/**
 * Facet for registring where to show CodeIntel tooltips.
 */
export const showTooltip: Facet<CodeIntelTooltip> = Facet.define<CodeIntelTooltip>({
    enables: self => [
        createLoaderExtension({
            input(state) {
                return state.facet(self)
            },
            create(tooltip) {
                return isObservable(tooltip.source)
                    ? new TooltipLoadingState(tooltip, Status.PENDING)
                    : new TooltipLoadingState(tooltip, Status.DONE, tooltip.source)
            },
            load(value) {
                return value.codeIntelTooltip.source as Observable<TooltipWithEnd | null>
            },
            provide: self => [
                uniqueTooltips.computeN([self], state => state.field(self).map(tooltip => tooltip.tooltip)),
            ],
        }),
    ],
})
