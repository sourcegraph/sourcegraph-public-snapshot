import { Decoration, EditorView, WidgetType } from '@codemirror/view'

class FocusCodeEditorShortcutWidget extends WidgetType {
    constructor() {
        super()
    }

    toDOM() {
        const element = document.createElement('kbd')
        element.classList.add('shortcut')
        element.innerText = 'C'
        return element
    }
}

/**
 * Extension adding focus code editor shortcut label to the end of the first line.
 */
export const focusCodeEditorShortcutLabel = [
    EditorView.decorations.compute([], state =>
        Decoration.set([
            Decoration.line({ class: 'position-relative' }).range(state.doc.line(1).from),
            Decoration.widget({
                widget: new FocusCodeEditorShortcutWidget(),
            }).range(state.doc.line(1).to),
        ])
    ),
    EditorView.theme({
        '.shortcut': {
            position: 'absolute',
            right: '0.25rem',
            display: 'inline-flex',
            alignItems: 'center',
        },
    }),
]
