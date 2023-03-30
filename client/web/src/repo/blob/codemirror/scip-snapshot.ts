import { Extension } from '@codemirror/state'
import { Decoration, EditorView, WidgetType } from '@codemirror/view'

class SCIPSnapshotDecorations extends WidgetType {
    constructor(private line: string) {
        super()
    }

    toDOM(view: EditorView): HTMLElement {
        const span = document.createElement('div')
        span.style.color = 'grey'
        span.innerText = this.line
        return span
    }
}

export const scipSnapshot = (data?: { offset: number; data: string }[]): Extension => {
    const widgets = data?.map(line => {
        console.log(line.data, line.offset)
        return Decoration.widget({
            widget: new SCIPSnapshotDecorations(line.data),
            block: true,
        }).range(line.offset, line.offset)
    })

    if (!widgets) return []

    return [EditorView.decorations.of(Decoration.set(widgets))]
}
