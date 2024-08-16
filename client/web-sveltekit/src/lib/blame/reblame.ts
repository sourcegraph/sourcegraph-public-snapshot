import { EditorView, GutterMarker } from '@codemirror/view'

import type { BlameHunk } from '$lib/web'

import ReblameMarkerComponent from './ReblameMarker.svelte'

export class ReblameMarker extends GutterMarker {
    private marker: ReblameMarkerComponent | null = null

    // hunk can be undefined if when the data is not available yet
    constructor(private line: number, private hunk: BlameHunk) {
        super()
    }

    public eq(other: ReblameMarker): boolean {
        // Only consider two markers with the same line equal if
        // hunk data is available. Otherwise the marker won't be
        // update/recreated as new data becomes available.
        return this.line === other.line
    }

    public toDOM(_view: EditorView): Node {
        const dom = document.createElement('div')
        dom.style.height = '100%'
        if (this.line !== 1) {
            dom.classList.add('sg-blame-border-top')
        }

        if (this.hunk.commit.previous) {
            this.marker = new ReblameMarkerComponent({
                target: dom,
                props: {
                    hunk: this.hunk,
                },
            })
        }
        return dom
    }

    public destroy(): void {
        this.marker?.$destroy()
    }
}
