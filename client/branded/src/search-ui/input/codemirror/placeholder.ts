/**
 * This is an adaption of the built-in CodeMirror placeholder to make it
 * configurable when the placeholder should be shown or not.
 */
import { type EditorState, type Extension, Facet } from '@codemirror/state'
import {
    Decoration,
    type DecorationSet,
    type EditorView,
    ViewPlugin,
    type ViewUpdate,
    WidgetType,
} from '@codemirror/view'

class Placeholder extends WidgetType {
    constructor(private readonly content: string) {
        super()
    }

    public toDOM(): HTMLElement {
        const wrap = document.createElement('span')
        wrap.className = 'cm-placeholder'
        wrap.style.pointerEvents = 'none'
        wrap.setAttribute('aria-hidden', 'true')
        wrap.append(document.createTextNode(this.content))
        return wrap
    }

    public ignoreEvent(): boolean {
        return false
    }
}

function showWhenEmpty(state: EditorState): boolean {
    return state.doc.length === 0
}

interface PlaceholderConfig {
    content: string
    show?: (state: EditorState) => boolean
}

export const placeholderConfig = Facet.define<PlaceholderConfig, Required<PlaceholderConfig>>({
    combine(configs) {
        // Keep highest priority config
        return configs.length > 0 ? { show: showWhenEmpty, ...configs[0] } : { content: '', show: showWhenEmpty }
    },
    enables: facet =>
        ViewPlugin.fromClass(
            class {
                private placeholderDecoration: Decoration
                public decorations: DecorationSet

                constructor(view: EditorView) {
                    const config = view.state.facet(facet)
                    this.placeholderDecoration = this.createWidget(config.content)
                    this.decorations = this.createDecorationSet(view.state, config)
                }

                public update(update: ViewUpdate): void {
                    let updateDecorations = update.docChanged || update.selectionSet

                    const config = update.view.state.facet(facet)
                    if (config !== update.startState.facet(facet)) {
                        this.placeholderDecoration = this.createWidget(config.content)
                        updateDecorations = true
                    }
                    if (updateDecorations) {
                        this.decorations = this.createDecorationSet(update.view.state, config)
                    }
                }

                private createWidget(content: string): Decoration {
                    return Decoration.widget({ widget: new Placeholder(content), side: 1 })
                }

                private createDecorationSet(state: EditorState, config: Required<PlaceholderConfig>): DecorationSet {
                    return config.show(state)
                        ? Decoration.set([this.placeholderDecoration.range(state.doc.length)])
                        : Decoration.none
                }
            },
            { decorations: plugin => plugin.decorations }
        ),
})

/**
 * Extension that shows a placeholder when the provided condition is met. By
 * default it will show the placeholder when the document is empty.
 */
export function placeholder(content: string, show: (state: EditorState) => boolean = showWhenEmpty): Extension {
    return placeholderConfig.of({ content, show })
}
