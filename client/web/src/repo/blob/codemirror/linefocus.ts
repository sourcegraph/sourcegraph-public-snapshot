import { Extension, RangeSetBuilder } from '@codemirror/state'
import { EditorView, Decoration, PluginValue, ViewUpdate, ViewPlugin } from '@codemirror/view'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import { SelectedLineRange, selectedLines, setSelectedLines } from './linenumbers'

class LineFocus implements PluginValue {
    private lastSelectedLine: number | null = null
    private lastFocusedLine: number
    private subscription: Subscription

    constructor(private view: EditorView, initialLine: number, onSelection: (range: SelectedLineRange) => void) {
        this.lastFocusedLine = initialLine

        // Listen to 'ArrowUp/Down' key events to focus the next/previous line.
        this.subscription = fromEvent<KeyboardEvent>(view.dom, 'keydown')
            .pipe(filter(event => event.key === 'ArrowUp' || event.key === 'ArrowDown'))
            .subscribe(event => {
                event.preventDefault()

                const nextLine =
                    event.key === 'ArrowUp'
                        ? Math.max(1, this.lastFocusedLine - 1)
                        : Math.min(this.lastFocusedLine + 1, view.state.doc.lines)

                this.focusLine(nextLine)
            })

        // Listen to focus events within CodeMirror to keep the focused line in sync.
        this.subscription.add(
            fromEvent<FocusEvent>(view.dom, 'focusin').subscribe(event => {
                const currentFocus = event.target as HTMLElement | null

                if (currentFocus) {
                    const nearestLine = view.state.doc.lineAt(view.posAtDOM(currentFocus))
                    this.lastFocusedLine = nearestLine.number
                }
            })
        )

        // Listen to 'Enter' key events to allow the user to update the selected line.
        this.subscription.add(
            fromEvent<KeyboardEvent>(view.dom, 'keydown')
                .pipe(filter(event => event.key === 'Enter'))
                .subscribe(event => {
                    const isLink = event.target instanceof HTMLAnchorElement
                    if (!isLink && this.lastFocusedLine) {
                        view.dispatch({
                            effects: setSelectedLines.of({ line: this.lastFocusedLine }),
                        })
                        onSelection({ line: this.lastFocusedLine })
                    }
                })
        )
    }

    public update(update: ViewUpdate): void {
        const currentSelectedLine = update.state.field(selectedLines)?.line

        // If the selected line has changed, focus the new line.
        if (currentSelectedLine && this.lastSelectedLine !== currentSelectedLine) {
            this.lastSelectedLine = currentSelectedLine
            return this.focusLine(currentSelectedLine)
        }

        // Otherwise, we need to ensure the last focused line remains in focus.
        // We need to do this on every update, as any update could potentially cause us to lose focus.
        // Note: We need to avoid triggering on `focusChanged` so we don't end up with a focus trap.
        if (this.lastFocusedLine && !update.focusChanged) {
            return this.focusLine(this.lastFocusedLine)
        }
    }

    public destroy(): void {
        this.subscription.unsubscribe()
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

interface FocusableLinesConfig {
    initialLine?: number
    onSelection: (range: SelectedLineRange) => void
}

const focusableLineDecoration = Decoration.line({ attributes: { tabIndex: '-1' } })

export function focusableLines({ initialLine = 1, onSelection }: FocusableLinesConfig): Extension {
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
        ViewPlugin.define(view => new LineFocus(view, initialLine, onSelection)),
    ]
}
