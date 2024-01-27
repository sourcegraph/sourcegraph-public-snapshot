/**
 * This provides CodeMirror extension for highlighting a static set of ranges.
 */
import { type Extension, StateEffect, EditorState, StateField, Facet } from '@codemirror/state'
import { Decoration, EditorView, showPanel, Panel } from '@codemirror/view'
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

const staticHighlightDecoration = Decoration.mark({ class: 'cm-sg-static-highlight' })
const staticHighlightSelectedDecoration = Decoration.mark({ class: 'cm-sg-static-highlight-selected' })

const staticHighlightState = StateField.define<{ from: number; to: number; selected: boolean }[]>({
    create: () => [],
    update: (ranges, tr) => {
        // When the selection changes, update the selection
        const selectedRange = tr.selection?.main
        if (selectedRange === undefined) {
            return ranges
        }
        return ranges.map(range => ({
            from: range.from,
            to: range.to,
            selected: range.from === selectedRange.from && range.to === selectedRange.to,
        }))
    },
    provide: f =>
        EditorView.decorations.from(f, ranges =>
            Decoration.set(
                ranges.map(({ selected, from, to }) =>
                    selected
                        ? staticHighlightSelectedDecoration.range(from, to)
                        : staticHighlightDecoration.range(from, to)
                )
            )
        ),
})

const staticHighlightTheme = EditorView.theme({
    '.cm-sg-static-highlight': {
        backgroundColor: 'var(--mark-bg)',
    },
    '.cm-sg-static-highlight-selected': {
        backgroundColor: 'var(--oc-orange-3)',
    },
})

export function staticHighlights(navigate: NavigateFunction, ranges: Range[]): Extension {
    const facet = Facet.define<Range[], Range[]>({
        combine: ranges => ranges.flat(),
        enables: () => [
            staticHighlightTheme,
            staticHighlightState.init(state =>
                ranges.map(range => ({
                    from: toCodeMirrorLocation(view.state, range.start),
                    to: toCodeMirrorLocation(view.state, range.end),
                }))
            ),
            showPanel.of(view => new StaticHighlightsPanel(view, navigate, ranges)),
        ],
    })
    return facet.of(ranges)
}

function toCodeMirrorLocation(state: EditorState, location: Location): number {
    // Codemirror expects 1-based line numbers
    return state.doc.line(location.line + 1).from + location.column
}

export const STATIC_HIGHLIGHTS_CONTAINER_ID = 'blob-search-container'

class StaticHighlightsPanel implements Panel {
    public dom: HTMLElement
    public top = true

    private root: Root | null = null

    constructor(private view: EditorView, private navigate: NavigateFunction) {
        this.dom = createElement('div', {
            className: classNames('cm-sg-search-container', styles.root),
            id: STATIC_HIGHLIGHTS_CONTAINER_ID,
        })
    }

    private get ranges(): { from: number; to: number; selected: boolean }[] {
        return this.view.state.field(staticHighlightState)
    }

    public update(): void {}

    public mount(): void {
        this.render()
        this.navigateTo(this.ranges[0])
    }

    private navigateTo(target: { from: number; to: number }) {
        this.view.dispatch({
            selection: { anchor: target.from, head: target.to },
            effects: [
                EditorView.scrollIntoView(target.from, {
                    y: 'nearest',
                    yMargin: this.view.dom.getBoundingClientRect().height / 3,
                }),
            ],
        })
    }

    private navigatePrevious() {
        const currentSelection = this.view.state.selection.main
        let previous: { from: number; to: number } | undefined
        for (const range of this.ranges) {
            previous = range
            if (range.from >= currentSelection.from) {
                break
            }
        }
        this.navigateTo(previous)
    }

    private navigateNext() {
        const currentSelection = this.view.state.selection.main
        let next: { from: number; to: number } | undefined
        for (const range of this.ranges) {
            next = range
            if (range.from > currentSelection.from) {
                break
            }
        }
        this.navigateTo(next)
    }

    private render(): void {
        if (!this.root) {
            this.root = createRoot(this.dom)
        }

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
                            onClick={this.navigatePrevious.bind(this)}
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
                            onClick={this.navigateNext.bind(this)}
                            data-testid="blob-view-static-next"
                            aria-label="next result"
                        >
                            <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                        </Button>
                    </div>
                )}
                <Text className="cm-search-results m-0 small">
                    {totalMatches} {pluralize('result', totalMatches)}
                </Text>
            </CodeMirrorContainer>
        )
    }
}
