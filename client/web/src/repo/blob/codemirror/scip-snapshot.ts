import type { Extension } from '@codemirror/state'
import { Decoration, EditorView, WidgetType } from '@codemirror/view'

class SCIPSnapshotDecorations extends WidgetType {
    constructor(private line: string, private additional: string[] | null) {
        super()
    }

    public toDOM(view: EditorView): HTMLElement {
        const div = document.createElement('div')
        const lineDiv = document.createElement('div')
        lineDiv.style.color = 'grey'
        lineDiv.innerText = this.line
        div.append(lineDiv)
        for (const extra of this.additional || []) {
            const extraDiv = document.createElement('div')
            extraDiv.style.color = 'grey'
            extraDiv.innerText = extra
            div.append(extraDiv)
        }
        return div
    }
}

export const scipSnapshot = (
    blob: string,
    data?: { offset: number; data: string; additional: string[] | null }[] | null
): Extension =>
    data
        ? [
              EditorView.decorations.of(
                  Decoration.set(
                      data.map(line =>
                          Decoration.widget({
                              widget: new SCIPSnapshotDecorations(line.data, line.additional),
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
