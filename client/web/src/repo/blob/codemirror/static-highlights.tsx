/**
 * This provides CodeMirror extension for highlighting a static set of ranges.
 */
import { type Extension, StateEffect, EditorState } from '@codemirror/state'
import { Decoration, EditorView, showPanel, Panel, ViewUpdate, ViewPlugin } from '@codemirror/view'
import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
import { createRoot, type Root } from 'react-dom/client'
import type { NavigateFunction } from 'react-router-dom'

import { pluralize } from '@sourcegraph/common'
import { Button, Icon, Text } from '@sourcegraph/wildcard'

import { createElement } from '../../../util/dom'

import { CodeMirrorContainer } from './react-interop'

import styles from './search.module.scss'

export interface Range {
    start: Location
    end: Location
}

export interface Location {
    // A zero-based line number
    line: number
    // A zero-based column number
    column: number
}

const staticHighlightDecoration = Decoration.mark({ class: 'cm-sg-staticSelection' })

export function staticHighlights(navigate: NavigateFunction, ranges: Range[]): Extension {
    return [
        EditorView.decorations.compute(['doc'], state =>
            Decoration.set(
                ranges.map(range =>
                    staticHighlightDecoration.range(
                        toCodeMirrorLocation(state, range.start),
                        toCodeMirrorLocation(state, range.end)
                    )
                ),
                true
            )
        ),
        EditorView.theme({
            '.cm-sg-staticSelection': {
                backgroundColor: 'var(--mark-bg)',
            },
        }),
        showPanel.of(view => new StaticHighlightsPanel(view, navigate, ranges)),
    ]
}

function toCodeMirrorLocation(state: EditorState, location: Location): number {
    // Codemirror expects 1-based line numbers
    return state.doc.line(location.line + 1).from + location.column
}

export const STATIC_HIGHLIGHTS_CONTAINER_ID = 'blob-search-container'
const navigateSelection = StateEffect.define<number>()

class StaticHighlightsPanel implements Panel {
    public dom: HTMLElement
    public top = true

    // Currently selected 0-based match index.
    private selected: number = 0
    private root: Root | null = null

    constructor(private view: EditorView, private navigate: NavigateFunction, private ranges: Range[]) {
        this.dom = createElement('div', {
            className: classNames('cm-sg-search-container', styles.root),
            id: STATIC_HIGHLIGHTS_CONTAINER_ID,
        })
    }

    public update(update: ViewUpdate): void {
        let diff = 0
        for (const tr of update.transactions) {
            for (const effect of tr.effects) {
                if (effect.is(navigateSelection)) {
                    diff += effect.value
                }
            }
        }
        const newSelected = (this.ranges.length + diff) % this.ranges.length
        if (newSelected !== this.selected) {
            this.selected = newSelected
            this.render()
            this.scrollToSelection() // Is it okay to dispatch in an update?
        }
    }

    public mount(): void {
        this.render()
    }

    public destroy(): void { }

    private scrollToSelection(): void {
        const selectedRange = this.ranges[this.selected]
        const codemirrorLocation =
            this.view.state.doc.line(selectedRange.start.line + 1).from + selectedRange.start.column
        this.view.dispatch({
            effects: EditorView.scrollIntoView(codemirrorLocation, {
                y: 'nearest',
                yMargin: this.view.dom.getBoundingClientRect().height / 3,
            }),
        })
    }

    private navigatePrevious() {
        this.view.dispatch({ effects: [navigateSelection.of(-1)] })
    }

    private navigateNext() {
        this.view.dispatch({ effects: [navigateSelection.of(1)] })
    }

    private render(): void {
        if (!this.root) {
            this.root = createRoot(this.dom)
        }
        const totalMatches = this.ranges.length

        this.root.render(
            <CodeMirrorContainer navigate={this.navigate}>
                {totalMatches > 1 && (
                    <div>
                        <Button
                            className={classNames(styles.bgroupLeft, 'p-1')}
                            type="button"
                            size="sm"
                            outline={true}
                            variant="secondary"
                            onClick={this.navigatePrevious}
                            data-testid="blob-view-static-previous"
                            aria-label="previous result"
                        >
                            <Icon svgPath={mdiChevronLeft} aria-hidden={true} />
                        </Button>

                        <Button
                            className={classNames(styles.bgroupRight, 'p-1')}
                            type="button"
                            size="sm"
                            outline={true}
                            variant="secondary"
                            onClick={this.navigateNext}
                            data-testid="blob-view-static-next"
                            aria-label="next result"
                        >
                            <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                        </Button>
                    </div>
                )}
                <Text className="cm-search-results m-0 small">
                    {`${this.selected} of `}
                    {totalMatches} {pluralize('result', totalMatches)}
                </Text>
            </CodeMirrorContainer>
        )
    }
}
