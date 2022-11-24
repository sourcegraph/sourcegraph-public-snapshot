import { Extension } from '@codemirror/state'
import { EditorView, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { StandardSuggestionSource } from './completion'
import Suggestions from './Suggestions.svelte'

type Item = { type: 'completion' }

class SuggestionView {
    private instance: Suggestions
    private root: HTMLElement

    constructor(public view: EditorView, public parent: HTMLDivElement) {
        this.root = document.createElement('div')
        this.instance = new Suggestions({
            target: parent,
            props: {
                view: this.view,
                parent,
            },
        })
        this.view.dom.appendChild(this.root)
    }

    update(update: ViewUpdate): void {
        if (update.focusChanged) {
            if (update.view.hasFocus) {
                this.instance.show()
            } else {
                this.instance.hide()
            }
        }
    }

    destroy() {
        this.instance.$destroy()
        this.root.remove()
    }
}

export const suggestions = (parent: HTMLDivElement, sources: StandardSuggestionSource[]): Extension => {
    return [ViewPlugin.define(view => new SuggestionView(view, parent))]
}
