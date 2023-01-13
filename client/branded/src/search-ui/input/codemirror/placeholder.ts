/**
 * This is an adaption of the built-in CodeMirror placeholder to make it
 * configurable when the placeholder should be shown or not.
 */
import { EditorState, Extension } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, ViewPlugin, ViewUpdate, WidgetType } from '@codemirror/view'

class Placeholder extends WidgetType {
    constructor(readonly content: string) {
        super()
    }

    toDOM() {
        let wrap = document.createElement('span')
        wrap.className = 'cm-placeholder'
        wrap.style.pointerEvents = 'none'
        wrap.setAttribute('aria-hidden', 'true')
        wrap.appendChild(document.createTextNode(this.content))
        return wrap
    }

    ignoreEvent() {
        return false
    }
}

function showWhenEmpty(state: EditorState): boolean {
    return state.doc.length === 0
}

/**
 * Extension that shows a placeholder when the provided condition is met. By
 * default it will show the placeholder when the document is empty.
 */
export function placeholder(content: string, show: (state: EditorState) => boolean = showWhenEmpty): Extension {
    return ViewPlugin.fromClass(
        class {
            placeholderDecoration: Decoration
            decorations: DecorationSet

            constructor(readonly view: EditorView) {
                this.placeholderDecoration = Decoration.widget({ widget: new Placeholder(content), side: 1 })
                this.decorations = this.createDecorationSet(view.state)
            }

            update(update: ViewUpdate) {
                if (update.docChanged || update.selectionSet) {
                    this.decorations = this.createDecorationSet(update.view.state)
                }
            }

            private createDecorationSet(state: EditorState): DecorationSet {
                return show(state)
                    ? Decoration.set([this.placeholderDecoration.range(state.doc.length)])
                    : Decoration.none
            }
        },
        { decorations: v => v.decorations }
    )
}
