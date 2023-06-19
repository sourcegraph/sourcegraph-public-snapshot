import { Extension } from '@codemirror/state'
import { Decoration, EditorView, WidgetType } from '@codemirror/view'

class SCIPSnapshotDecorations extends WidgetType {
    constructor(private line: string) {
        super()
    }

    public toDOM(view: EditorView): HTMLElement {
        const span = document.createElement('div')
        span.style.color = 'grey'
        span.innerText = this.line
        return span
    }
}

export const scipSnapshot = (blob: string, data?: { offset: number; data: string }[] | null): Extension =>
    data
        ? [
              EditorView.decorations.of(
                  Decoration.set(
                      data.map(line =>
                          Decoration.widget({
                              widget: new SCIPSnapshotDecorations(line.data),
                              block: true,
                          }).range(
                              // If the offset is beyond the document, we have to bring it back one
                              // so codemirror will render them. This only looks correct when there
                              // are no lines after it, which is the case for the final line, otherwise
                              // offsetting by -1 gives weird extra newlines
                              line.offset - (line.offset > blob.length ? 1 : 0),
                              line.offset - (line.offset > blob.length ? 1 : 0)
                          )
                      )
                  )
              ),
          ]
        : []
