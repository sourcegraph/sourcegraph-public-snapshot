import { Extension, Range as RangesetRange, RangeSet } from '@codemirror/state'
import { EditorView, Decoration, WidgetType } from '@codemirror/view'

import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

import styles from './comment-annotations.module.scss'

class FunctionInfoWidget extends WidgetType {
    constructor(public readonly title: string, public readonly tabSize: number) {
        super()
    }

    // eslint-disable-next-line id-length
    public eq(other: FunctionInfoWidget): boolean {
        return other.title === this.title && other.tabSize === this.tabSize
    }

    public toDOM(): HTMLElement {
        const wrap = document.createElement('div')
        wrap.className = styles.comment
        wrap.innerHTML = this.title
        wrap.style.paddingLeft = `${this.tabSize * 8}px` // 8px is the default width of a tab character
        return wrap
    }

    public ignoreEvent(): boolean {
        return false
    }
}

export type Comments = Record<number, string>

export const [comments, setComments] = createUpdateableField<Comments | null>(null, field => [
    EditorView.decorations.compute([field], state => {
        const comments = state.field(field, false) ?? null

        if (comments === null) {
            return Decoration.none
        }

        const widgets: RangesetRange<Decoration>[] = []

        // visible lines
        const fromLine = 40
        const toLine = 100

        // Iterate over lines and add widgets
        for (let line = fromLine; line <= toLine; line++) {
            const comment = comments[line]

            if (!comment) {
                continue
            }

            const docLine = state.doc.line(line)
            const lineContent = docLine.text
            const tabSize = lineContent.search(/\S|$/)

            const widget = Decoration.widget({
                widget: new FunctionInfoWidget(comment, tabSize),
                block: true,
            }).range(docLine.from)

            widgets.push(widget)
        }

        return widgets.length > 0 ? RangeSet.of(widgets) : Decoration.none
    }),
])

export function annotateWithComments(newComments?: Comments): Extension {
    return [comments.init(() => newComments ?? null)]
}
