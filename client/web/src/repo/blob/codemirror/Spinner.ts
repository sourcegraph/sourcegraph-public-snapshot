import { EditorView, getTooltip, Tooltip } from '@codemirror/view'
import { setHoverEffect } from './textdocument-hover'

export class Spinner {
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
            this.view.dispatch({ effects: setHoverEffect.of(this.tooltip) })
        }
    }
    public stop(): void {
        if (this.tooltip && getTooltip(this.view, this.tooltip)) {
            this.view.dispatch({ effects: setHoverEffect.of(null) })
        }
    }
}
