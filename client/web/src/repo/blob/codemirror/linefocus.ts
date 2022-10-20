import { Extension, RangeSetBuilder } from '@codemirror/state'
import { EditorView, Decoration, PluginValue, ViewUpdate, ViewPlugin } from '@codemirror/view'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import { selectedLines } from './linenumbers'

class LineFocus implements PluginValue {
    private lastSelectedLine: number | null = null
    private lastFocusedLine: number
    private keySubscription: Subscription
    private focusSubscription: Subscription

    constructor(private readonly view: EditorView, initialLine: number) {
        this.lastFocusedLine = initialLine
        this.keySubscription = fromEvent<KeyboardEvent>(view.dom, 'keydown')
            .pipe(filter(event => event.key === 'ArrowUp' || event.key === 'ArrowDown'))
            .subscribe(event => {
                event.preventDefault()

                const nextLine =
                    event.key === 'ArrowUp'
                        ? Math.max(1, this.lastFocusedLine - 1)
                        : Math.min(this.lastFocusedLine + 1, view.state.doc.lines)

                this.focusLine(nextLine)
            })

        this.focusSubscription = fromEvent<FocusEvent>(view.dom, 'focusin').subscribe(event => {
            const currentFocus = event.target as HTMLElement | null

            if (currentFocus) {
                const nearestLine = view.state.doc.lineAt(view.posAtDOM(currentFocus))
                this.lastFocusedLine = nearestLine.number
            }
        })
    }

    public update(update: ViewUpdate): void {
        const currentSelectedLine = update.state.field(selectedLines)?.line
        if (currentSelectedLine && this.lastSelectedLine !== currentSelectedLine) {
            this.lastSelectedLine = currentSelectedLine
            this.focusLine(currentSelectedLine)
        } else {
            this.focusLine(this.lastFocusedLine)
        }
    }

    public destroy(): void {
        this.keySubscription.unsubscribe()
        this.focusSubscription.unsubscribe()
    }

    private focusLine(lineNumber: number): void {
        const nextLine = this.view.state.doc.line(lineNumber)
        const nextLineElement = this.view.domAtPos(nextLine.from).node as HTMLElement | null
        if (nextLineElement) {
            this.lastFocusedLine = lineNumber
            window.requestAnimationFrame(() => {
                nextLineElement.focus()
            })
        }
    }
}

const focusableLineDecoration = Decoration.line({ attributes: { tabIndex: '-1' } })

export function focusableLines(initialLine = 1): Extension {
    return [
        EditorView.decorations.compute([], state => {
            const to = state.doc.lines
            const builder = new RangeSetBuilder<Decoration>()

            for (let lineNumber = 1; lineNumber <= to; lineNumber++) {
                const from = state.doc.line(lineNumber).from
                builder.add(from, from, focusableLineDecoration)
            }

            return builder.finish()
        }),
        ViewPlugin.define(view => new LineFocus(view, initialLine)),
    ]
}
