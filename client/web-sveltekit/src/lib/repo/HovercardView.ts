import type { EditorView, TooltipView } from '@codemirror/view'

import type { TooltipViewOptions } from '$lib/web'

import Hovercard from './HovercardView.svelte'

export class HovercardView implements TooltipView {
    public readonly dom: HTMLElement
    private readonly hovercard: Hovercard

    constructor(
        // TODO: Add support for pinned tooltips
        _view: EditorView,
        _tokenRange: TooltipViewOptions['token'],
        hovercardData: TooltipViewOptions['hovercardData']
    ) {
        this.dom = document.createElement('div')
        this.hovercard = new Hovercard({
            target: this.dom,
            props: {
                hovercardData,
            },
        })
    }

    public destroy(): void {
        this.hovercard.$destroy()
    }
}
