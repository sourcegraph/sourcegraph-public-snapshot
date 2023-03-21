import { Decoration, EditorView, WidgetType } from '@codemirror/view'

class Banana extends WidgetType {
    constructor(private line: string) {
        super()
    }

    toDOM(view: EditorView): HTMLElement {
        const span = document.createElement('div')
        span.style.color = 'grey'
        // span.style.
        span.innerText = this.line
        return span
    }
}

export const scipSnapshot = (data?: { offset: number; data: string }[]) => {
    const widgets = data?.map(line => {
        console.log(line.data, line.offset)
        return Decoration.widget({
            widget: new Banana(line.data),
            block: true,
        }).range(line.offset, line.offset)
    })!!

    return EditorView.decorations.of(Decoration.set(widgets))
}
