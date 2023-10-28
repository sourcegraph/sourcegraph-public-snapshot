import {
    EditorState,
    Extension,
    Facet,
    RangeSetBuilder,
    StateEffect,
    StateField,
    Transaction,
    TransactionSpec,
} from '@codemirror/state'
import {
    Tooltip,
    showTooltip as showCodeMirrorTooltip,
    EditorView,
    PluginValue,
    ViewUpdate,
    getTooltip,
    TooltipView,
    ViewPlugin,
    Decoration,
} from '@codemirror/view'

import { LoadingTooltip } from '../tooltips/LoadingTooltip'

import { getCodeIntelAPI, CodeIntelTooltipPosition } from './api'
import { selectedToken } from './token-selection'

const setTooltipRange = StateEffect.define<{ key: unknown; range: Range | null }>()
const setTooltip = StateEffect.define<{ state: TooltipState; tooltip: Tooltip | null }>()
const showTooltipFor = StateEffect.define<{ key: unknown; showLoadingTooltip?: boolean }>()
const hideTooltipFor = StateEffect.define<{ range?: Range; key?: unknown }>()

type Range = CodeIntelTooltipPosition['range']

enum Status {
    HIDDEN,
    PENDING,
    VISIBLE,
}

function isEqual(rangeA: Range, rangeB: Range) {
    return rangeA.to === rangeB.to && rangeA.from === rangeB.from
}

export class TooltipState {
    public visible: boolean

    constructor(
        public position: CodeIntelTooltipPosition,
        public status: Status,
        public tooltip: Tooltip | null = null
    ) {
        this.visible = !!tooltip && (status === Status.VISIBLE || status === Status.PENDING)
    }

    update(tr: Transaction): TooltipState {
        let state: TooltipState = this

        for (const effect of tr.effects) {
            if (effect.is(setTooltip) && effect.value.state === this && state.status !== Status.HIDDEN) {
                state = new TooltipState(
                    state.position,
                    effect.value.tooltip ? Status.VISIBLE : Status.HIDDEN,
                    effect.value.tooltip ?? state.tooltip
                )
            } else if (effect.is(showTooltipFor) && !this.visible && state.position.key === effect.value.key) {
                state = new TooltipState(
                    state.position,
                    state.tooltip ? Status.VISIBLE : Status.PENDING,
                    state.tooltip ??
                        (effect.value.showLoadingTooltip ? new LoadingTooltip(state.position.range.from) : null)
                )
            } else if (
                effect.is(hideTooltipFor) &&
                (effect.value.key === state.position.key ||
                    (effect.value.range && isEqual(effect.value.range, state.position.range))) &&
                state.status !== Status.HIDDEN
            ) {
                state = new TooltipState(state.position, Status.HIDDEN, state.tooltip)
            }
        }

        return state
    }
}

class CodeIntelState {
    constructor(public tooltips: TooltipState[]) {}

    static create(positions: readonly CodeIntelTooltipPosition[]) {
        return new CodeIntelState(positions.map(position => new TooltipState(position, Status.PENDING)))
    }

    update(tr: Transaction): CodeIntelState {
        let state: CodeIntelState = this
        const newPositions = tr.state.facet(showTooltips)
        if (newPositions !== tr.startState.facet(showTooltips)) {
            state = state.addOrRemoveTooltips(newPositions)
        }

        // Update tooltips
        const newTooltips = state.tooltips.map(tooltip => tooltip.update(tr))
        if (
            newTooltips.length !== this.tooltips.length ||
            newTooltips.some((tooltip, i) => tooltip !== this.tooltips[i])
        ) {
            state = new CodeIntelState(newTooltips)
        }

        return state
    }

    private addOrRemoveTooltips(positions: readonly CodeIntelTooltipPosition[]): CodeIntelState {
        let changed = false
        let tooltips = this.tooltips.filter(tooltip => {
            const keep = positions.some(position => position === tooltip.position)
            changed = changed || !keep
            return keep
        })

        for (const position of positions) {
            if (!this.tooltips.some(tooltip => tooltip.position === position)) {
                changed = true
                tooltips.push(new TooltipState(position, Status.PENDING))
            }
        }

        return changed ? new CodeIntelState(tooltips) : this
    }

    pendingTooltips() {
        return this.tooltips.filter(tooltips => tooltips.status === Status.PENDING)
    }

    getTooltip(key: unknown): TooltipState | null {
        return this.tooltips.find(tooltip => tooltip.position.key === key) ?? null
    }
}

export function getTooltipState(state: EditorState, key: unknown): TooltipState | null {
    return state.field(tooltipLoadingState).getTooltip(key) ?? null
}

function isOffsetRange(offset: number, range: { from: number; to: number }): boolean {
    return range.from <= offset && offset <= range.to
}

export function hasTooltipAtOffset(state: EditorState, offset: number, key?: unknown) {
    return state
        .field(tooltipLoadingState)
        .tooltips.some(
            tooltip =>
                (!key || (tooltip.position.key === key && tooltip.visible)) &&
                isOffsetRange(offset, tooltip.position.range)
        )
}

export function getTooltipViewFor(view: EditorView, key: unknown): TooltipView | null {
    const tooltip = view.state
        .field(tooltipLoadingState)
        .tooltips.find(tooltip => tooltip.position.key === key)?.tooltip
    return tooltip ? getTooltip(view, tooltip) : null
}

export function showCodeIntelTooltipAtRange(
    view: EditorView,
    range: Range,
    key?: unknown | boolean,
    showLoadingTooltip?: boolean
): void {
    // TODO: preload data
    view.dispatch({
        effects: [
            setTooltipRange.of({ range, key }),
            showTooltipFor.of({ key, showLoadingTooltip: showLoadingTooltip }),
        ],
    })
}

export function hideTooltipForKey(key: unknown): TransactionSpec {
    return { effects: [hideTooltipFor.of({ key })] }
}

const tooltipTheme = EditorView.theme({
    '.cm-tooltip.sg-code-intel-hovercard': {
        border: 'unset',
    },
})

/**
 * {@link StateField} storing focused (selected), hovered and pinned {@link Occurrence}s and {@link Tooltip}s associate with them.
 */
const tooltipLoadingState = StateField.define<CodeIntelState>({
    create(state) {
        return CodeIntelState.create(state.facet(showTooltips))
    },
    update(value, transaction) {
        return value.update(transaction)
    },
    provide(self) {
        return [
            tooltipTheme,

            // View plugin responsible for fetching tooltip content
            ViewPlugin.fromClass(
                class TooltipLoader implements PluginValue {
                    private pendingTooltips = new Set<TooltipState>()

                    constructor(private view: EditorView) {
                        for (const tooltip of view.state.field(self).pendingTooltips()) {
                            this.loadTooltip(tooltip)
                        }
                    }

                    update(update: ViewUpdate) {
                        const state = update.state.field(self)
                        if (state.tooltips !== update.startState.field(self).tooltips) {
                            for (const tooltip of state.pendingTooltips()) {
                                if (!this.pendingTooltips.has(tooltip)) {
                                    this.loadTooltip(tooltip)
                                }
                            }
                        }
                    }

                    private async loadTooltip(tooltipState: TooltipState): Promise<void> {
                        this.pendingTooltips.add(tooltipState)
                        const tooltip = await getCodeIntelAPI(this.view.state).getHoverTooltip(
                            this.view.state,
                            tooltipState.position
                        )
                        this.view.dispatch({ effects: setTooltip.of({ state: tooltipState, tooltip }) })
                        this.pendingTooltips.delete(tooltipState)
                    }
                }
            ),
        ]
    },
})

const uniqueTooltips = Facet.define<TooltipState>({
    combine(values) {
        return [...values].sort((a, b) => a.position.range.from - b.position.range.from)
    },
    enables: self => [
        // Highlight tokens with tooltips
        EditorView.decorations.compute([self, selectedToken], state => {
            let decorations = new RangeSetBuilder<Decoration>()
            const tooltips = state.facet(self)
            const selectedRange = state.field(selectedToken)
            for (const {
                position: { range, key },
            } of tooltips) {
                // We shouldn't add/remove any decorations inside the selected token, because
                // that causes the node to be recreated and loosing focus, which breaks
                // token keyboard navigation.
                if (!selectedRange || !isEqual(range, selectedRange)) {
                    decorations.add(
                        range.from,
                        range.to,
                        Decoration.mark({ class: `selection-highlight selection-highlight-${key}` })
                    )
                }
            }

            return decorations.finish()
        }),

        // Show tooltips
        showCodeMirrorTooltip.computeN([self], state => state.facet(self).map(({ tooltip }) => tooltip)),
    ],
})

/**
 * Facet for registring where to show CodeIntel tooltips.
 */
export const showTooltips = Facet.define<CodeIntelTooltipPosition>({
    combine(values) {
        const seen = new Set<unknown>()
        return values.filter(value => {
            if (seen.has(value.key)) {
                return false
            }
            seen.add(value.key)
            return true
        })
    },
})

/**
 * Field for keeping track of dynamic tooltip positions (e.g. hover).
 */
const dynamicTooltips = StateField.define<CodeIntelTooltipPosition[]>({
    create() {
        return []
    },
    update(tooltips, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setTooltipRange)) {
                const newTooltip = effect.value
                if (tooltips.some(tooltip => tooltip.key === newTooltip.key)) {
                    tooltips = tooltips.filter(tooltip => tooltip.key !== newTooltip.key)
                }
                if (newTooltip.range !== null) {
                    tooltips = [...tooltips, newTooltip as CodeIntelTooltipPosition]
                }
            }
        }
        return tooltips
    },
    provide: self => showTooltips.computeN([self], state => state.field(self)),
})

export function tooltipsExtension(options: { tooltipPriority: unknown[] }): Extension {
    const { tooltipPriority } = options
    return [
        tooltipLoadingState,
        dynamicTooltips,
        uniqueTooltips.computeN([tooltipLoadingState], state => {
            const tooltips: Map<number, TooltipState> = new Map()
            for (const tooltip of Array.from(state.field(tooltipLoadingState).tooltips).sort(
                (a, b) => tooltipPriority.indexOf(a.position.key) - tooltipPriority.indexOf(b.position.key)
            )) {
                if (tooltip.visible && tooltip.tooltip && !tooltips.has(tooltip.position.range.from)) {
                    tooltips.set(tooltip.position.range.from, tooltip)
                }
            }

            return Array.from(tooltips.values())
        }),
    ]
}
