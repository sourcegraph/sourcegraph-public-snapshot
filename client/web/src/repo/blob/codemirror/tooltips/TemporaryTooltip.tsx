import { EditorView, getTooltip, Tooltip, TooltipView } from '@codemirror/view'

import * as sourcegraph from '@sourcegraph/extension-api-types'

import { closeHover, showHover } from '../token-selection/hover'

class TemporaryTooltip implements Tooltip {
    public readonly above = true
    constructor(
        public readonly message: string,
        public readonly pos: number,
        public readonly arrow: boolean | undefined
    ) {}
    public create(): TooltipView {
        const div = document.createElement('div')
        div.textContent = this.message
        return { dom: div }
    }
}

// Displays a simple tooltip that automatically hides after the provided timeout
export function showTemporaryTooltip(
    view: EditorView,
    message: string,
    position: sourcegraph.Position,
    clearTimeout: number,
    options?: {
        arrow?: boolean
    }
): void {
    const line = view.state.doc.line(position.line + 1)
    const pos = line.from + position.character + 1
    const tooltip = new TemporaryTooltip(message, pos, options?.arrow)
    showHover(view, tooltip)
    setTimeout(() => {
        if (getTooltip(view, tooltip)) {
            closeHover(view)
        }
    }, clearTimeout)
}
