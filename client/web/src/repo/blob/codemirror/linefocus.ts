import { Extension, RangeSetBuilder } from '@codemirror/state'
import { EditorView, Decoration, ViewPlugin, PluginValue, ViewUpdate } from '@codemirror/view'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import { SelectedLineRange, selectedLines, setSelectedLines } from './linenumbers'

interface FocusableLinesConfig {
    initialLine?: number
    onSelection: (range: SelectedLineRange) => void
}

class LineFocus implements PluginValue {
    private lastSelectedLine: number | null = null
    private lastFocusedLine: number
    private subscription: Subscription

    constructor(
        private readonly view: EditorView,
        initialLine: number,
        onSelection: (range: SelectedLineRange) => void
    ) {
        this.lastFocusedLine = initialLine
        this.subscription = fromEvent<KeyboardEvent>(view.dom, 'keydown')
            .pipe(filter(event => event.key === 'ArrowUp' || event.key === 'ArrowDown'))
            .subscribe(event => {
                event.preventDefault()

                const nextLine =
                    event.key === 'ArrowUp'
                        ? Math.max(1, this.lastFocusedLine - 1)
                        : Math.min(this.lastFocusedLine + 1, view.state.doc.lines)

                // Allow scroll behavior here so that the line is always visible
                this.focusLine({ line: nextLine, preventScroll: false })

                // Update the current selection to easily allow copying the current text
                view.dispatch({ selection: { anchor: view.state.doc.line(nextLine).from } })
            })

        this.subscription.add(
            fromEvent<KeyboardEvent>(view.dom, 'keydown')
                .pipe(filter(event => event.key === 'Enter'))
                .subscribe(event => {
                    const hasAction =
                        event.target instanceof HTMLAnchorElement || event.target instanceof HTMLButtonElement
                    if (!hasAction) {
                        view.dispatch({
                            effects: setSelectedLines.of({ line: this.lastFocusedLine }),
                            selection: { anchor: view.state.doc.line(this.lastFocusedLine).from },
                        })
                        onSelection({ line: this.lastFocusedLine })
                    }
                })
        )

        this.subscription.add(
            fromEvent<FocusEvent>(view.dom, 'focusin').subscribe(event => {
                if (event.target instanceof HTMLElement) {
                    const nearestLine = view.state.doc.lineAt(view.posAtDOM(event.target))
                    this.lastFocusedLine = nearestLine.number
                }
            })
        )
    }

    public update(update: ViewUpdate): void {
        const currentSelectedLine = update.state.field(selectedLines)?.line
        if (currentSelectedLine && this.lastSelectedLine !== currentSelectedLine) {
            this.lastSelectedLine = currentSelectedLine
            this.focusLine({ line: currentSelectedLine })
        } else {
            this.focusLine({ line: this.lastFocusedLine })
        }
    }

    public destroy(): void {
        this.subscription.unsubscribe()
    }

    private focusLine({ line, preventScroll = true }: { line: number; preventScroll?: boolean }): void {
        const nextLine = this.view.state.doc.line(line)
        const nextLineElement = this.view.domAtPos(nextLine.from).node
        if (nextLineElement instanceof HTMLElement) {
            this.lastFocusedLine = line
            window.requestAnimationFrame(() => {
                if (document.activeElement !== nextLineElement && !this.view.dom.contains(document.activeElement)) {
                    nextLineElement.focus({ preventScroll })
                }
            })
        }
    }
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
