import { StateEffect, StateField } from '@codemirror/state'
import type { EditorView, Tooltip, TooltipView } from '@codemirror/view'

import { showTooltip } from '../codeintel/tooltips'

export class TemporaryTooltip implements Tooltip {
    public readonly above = true
    public readonly pos: number
    public readonly end?: number

    constructor(
        public range: { from: number; to?: number },
        public readonly message: string,
        public readonly arrow: boolean | undefined
    ) {
        this.pos = range.from
        this.end = range.to
    }

    public create(): TooltipView {
        const div = document.createElement('div')
        div.classList.add('tmp-tooltip')
        div.textContent = this.message
        return {
            dom: div,
        }
    }
}

const setTooltip = StateEffect.define<Tooltip>()
const clearTooltip = StateEffect.define<Tooltip>()

export const temporaryTooltip = StateField.define<Tooltip[]>({
    create() {
        return []
    },
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setTooltip) && !value.includes(effect.value)) {
                value = [...value, effect.value]
            } else if (effect.is(clearTooltip) && value.includes(effect.value)) {
                value = value.slice()
                value.splice(value.indexOf(effect.value), 1)
            }
        }
        return value
    },
    provide: field => showTooltip.computeN([field], state => state.field(field)),
})

/**
 * Displays a simple tooltip that automatically hides after the provided timeout.
 */
export function showTemporaryTooltip(view: EditorView, message: string, offset: number, clearTimeout: number): void {
    const tooltip = new TemporaryTooltip({ from: offset }, message, true)
    view.dispatch({ effects: setTooltip.of(tooltip) })
    setTimeout(() => {
        view.dispatch({ effects: clearTooltip.of(tooltip) })
    }, clearTimeout)
}
