import { EditorView, getTooltip, Tooltip } from '@codemirror/view'

import { closeHover, showHover } from '../token-selection/hover'

/** Helper to display a "Loading..." CodeMirror tooltip. */
export class LoadingTooltip {
    private tooltip?: Tooltip
    constructor(private view: EditorView, pos: number | null) {
        if (pos) {
            this.tooltip = {
                pos,
                above: true,
                create() {
                    const dom = document.createElement('div')
                    dom.textContent = 'Loading...'
                    return { dom }
                },
            }
            showHover(this.view, this.tooltip)
        }
    }
    public stop(): void {
        if (this.tooltip && getTooltip(this.view, this.tooltip)) {
            closeHover(this.view)
        }
    }
}
