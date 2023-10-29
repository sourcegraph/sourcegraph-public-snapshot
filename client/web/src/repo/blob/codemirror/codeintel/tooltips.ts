import { Facet, RangeSetBuilder, StateEffect, StateField, Transaction } from '@codemirror/state'
import {
    Tooltip,
    showTooltip as showCodeMirrorTooltip,
    EditorView,
    PluginValue,
    ViewUpdate,
    ViewPlugin,
    Decoration,
} from '@codemirror/view'
import { Observable, Subscription, isObservable } from 'rxjs'

import { CodeIntelTooltipPosition } from './api'
import { codeIntelDecorations } from './decorations'

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

const tooltipTheme = EditorView.theme({
    '.cm-tooltip.sg-code-intel-hovercard': {
        border: 'unset',
    },
})

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
        tooltipTheme,
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

const setTooltip = StateEffect.define<{ source: Observable<TooltipWithEnd>; tooltip: TooltipWithEnd | null }>()

export class TooltipLoadingState {
    public visible: boolean

    constructor(
        public codeIntelTooltip: CodeIntelTooltip,
        public status: Status,
        public tooltip: TooltipWithEnd | null = null
    ) {
        this.visible = !!tooltip
    }

    update(tr: Transaction): TooltipLoadingState {
        let state: TooltipLoadingState = this

        for (const effect of tr.effects) {
            if (effect.is(setTooltip) && effect.value.source === this.codeIntelTooltip.source) {
                state = new TooltipLoadingState(state.codeIntelTooltip, Status.DONE, effect.value.tooltip)
            }
        }

        return state
    }
}

class CodeIntelState {
    constructor(public tooltips: TooltipLoadingState[]) {}

    static create(tooltips: readonly CodeIntelTooltip[]) {
        return new CodeIntelState(tooltips.map(tooltip => new TooltipLoadingState(tooltip, Status.PENDING)))
    }

    update(tr: Transaction): CodeIntelState {
        let state: CodeIntelState = this
        const newPositions = tr.state.facet(showTooltip)
        if (newPositions !== tr.startState.facet(showTooltip)) {
            state = state.addOrRemoveTooltips(newPositions)
        }

        // Update tooltips
        state = new CodeIntelState(state.tooltips.map(tooltip => tooltip.update(tr)))
        return this.eq(state) ? this : state
    }

    eq(other: CodeIntelState): boolean {
        return (
            this.tooltips.length === other.tooltips.length &&
            this.tooltips.every((tooltip, i) => tooltip === other.tooltips[i])
        )
    }

    private addOrRemoveTooltips(newTooltips: readonly CodeIntelTooltip[]): CodeIntelState {
        return new CodeIntelState(
            newTooltips.map(tooltip => {
                let seenAt = -1
                for (let i = 0; i < this.tooltips.length; i++) {
                    if (tooltip === this.tooltips[i].codeIntelTooltip) {
                        seenAt = i
                    }
                }
                return seenAt > -1
                    ? this.tooltips[seenAt]
                    : new TooltipLoadingState(
                          tooltip,
                          isObservable(tooltip.source) ? Status.PENDING : Status.DONE,
                          isObservable(tooltip.source) ? null : tooltip.source
                      )
            })
        )
    }

    dynamicSources(): Observable<TooltipWithEnd>[] {
        return this.tooltips
            .filter(tooltip => isObservable(tooltip.codeIntelTooltip.source))
            .map(tooltip => tooltip.codeIntelTooltip.source as Observable<TooltipWithEnd>)
    }
}

/**
 * {@link StateField} storing focused (selected), hovered and pinned {@link Occurrence}s and {@link Tooltip}s associate with them.
 */
const tooltipLoadingState = StateField.define<CodeIntelState>({
    create(state) {
        return CodeIntelState.create(state.facet(showTooltip))
    },
    update(value, transaction) {
        return value.update(transaction)
    },
    provide(self) {
        return [
            // View plugin responsible for fetching tooltip content
            ViewPlugin.fromClass(
                class TooltipLoader implements PluginValue {
                    private tooltips = new Map<Observable<Tooltip>, Subscription>()

                    constructor(private view: EditorView) {
                        for (const source of view.state.field(self).dynamicSources()) {
                            this.loadTooltip(source)
                        }
                    }

                    update(update: ViewUpdate) {
                        const loadingState = update.state.field(self)
                        if (loadingState !== update.startState.field(self)) {
                            const dynamicSources = loadingState.dynamicSources()
                            for (const [source, subscription] of this.tooltips) {
                                if (!dynamicSources.some(dynamicSource => dynamicSource === source)) {
                                    subscription.unsubscribe()
                                    this.tooltips.delete(source)
                                }
                            }

                            for (const source of dynamicSources) {
                                if (!this.tooltips.has(source)) {
                                    this.loadTooltip(source)
                                }
                            }
                        }
                    }

                    private async loadTooltip(source: Observable<TooltipWithEnd>): Promise<void> {
                        this.tooltips.set(
                            source,
                            source.subscribe(tooltip => {
                                this.view.dispatch({ effects: setTooltip.of({ source, tooltip }) })
                            })
                        )
                    }
                }
            ),

            uniqueTooltips.computeN([self], state => state.field(self).tooltips.map(tooltip => tooltip.tooltip)),
        ]
    },
})

/**
 * Facet for registring where to show CodeIntel tooltips.
 */
export const showTooltip: Facet<CodeIntelTooltip> = Facet.define<CodeIntelTooltip>({
    enables: [tooltipLoadingState],
})
