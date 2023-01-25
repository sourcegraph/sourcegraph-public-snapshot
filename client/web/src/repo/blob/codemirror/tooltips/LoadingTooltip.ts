import { Tooltip } from '@codemirror/view'

export class LoadingTooltip implements Tooltip {
    public readonly above = true

    constructor(public readonly pos: number) {}

    public create() {
        const dom = document.createElement('div')
        dom.classList.add('tmp-tooltip')
        dom.textContent = 'Loading...'
        return { dom }
    }
}
