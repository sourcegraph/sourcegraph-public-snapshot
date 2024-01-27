/**
 * This provides CodeMirror extension for highlighting a static set of ranges.
 */
import { type Extension, EditorState, StateField, Facet } from '@codemirror/state'
import { Decoration, EditorView, showPanel, Panel, ViewUpdate } from '@codemirror/view'
import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'
import { sortedIndexBy } from 'lodash'
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

interface HighlightedRange {
    from: number
    to: number
    selected: boolean
}

const staticHighlightDecoration = Decoration.mark({ class: 'cm-sg-static-highlight' })
const staticHighlightSelectedDecoration = Decoration.mark({ class: 'cm-sg-static-highlight-selected' })

const staticHighlightState = StateField.define<HighlightedRange[]>({
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
                ),
                true
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
    if (!ranges) {
        return []
    }
    const facet = Facet.define<Range[], Range[]>({
        combine: ranges => ranges.flat(),
        enables: () => [
            staticHighlightTheme,
            staticHighlightState.init(state =>
                ranges.map((range, i) => ({
                    selected: i === 0,
                    from: toCodeMirrorLocation(state, range.start),
                    to: toCodeMirrorLocation(state, range.end),
                }))
            ),
            EditorView.updateListener.of(update => {
                if (update.docChanged) {
                    const range = toCodeMirrorRange(update.state, ranges[0])
                    console.log({ range })
                    navigateTo(update.view, range)
                }
            }),
            showPanel.of(view => new StaticHighlightsPanel(view, navigate)),
        ],
    })
    return facet.of(ranges)
}

function toCodeMirrorRange(state: EditorState, range: Range): { from: number; to: number } {
    return {
        from: toCodeMirrorLocation(state, range.start),
        to: toCodeMirrorLocation(state, range.end),
    }
}

function toCodeMirrorLocation(state: EditorState, location: Location): number {
    // Codemirror expects 1-based line numbers
    return state.doc.line(location.line + 1).from + location.column
}

function findSelectedIdx(ranges: HighlightedRange[]): number | undefined {
    const selectedIdx = ranges.findIndex(range => range.selected)
    return selectedIdx === -1 ? undefined : selectedIdx
}

export const STATIC_HIGHLIGHTS_CONTAINER_ID = 'static-highlights-navigation-container'

class StaticHighlightsPanel implements Panel {
    public dom: HTMLElement
    public top = true

    private root: Root | null = null

    constructor(private view: EditorView, private navigate: NavigateFunction) {
        this.dom = createElement('div', {
            className: classNames('cm-sg-search-container', styles.root),
            id: STATIC_HIGHLIGHTS_CONTAINER_ID,
        })
        this.render(this.ranges)
    }

    private get ranges(): { from: number; to: number; selected: boolean }[] {
        return this.view.state.field(staticHighlightState)
    }

    public update(viewUpdate: ViewUpdate): void {
        const oldSelectedIdx = findSelectedIdx(viewUpdate.startState.field(staticHighlightState))
        const newSelectedIdx = findSelectedIdx(viewUpdate.state.field(staticHighlightState))
        if (newSelectedIdx !== oldSelectedIdx) {
            this.render(viewUpdate.state.field(staticHighlightState))
        }
    }

    private navigatePrevious() {
        const currentSelection = this.view.state.selection.main
        const idx = sortedIndexBy(
            this.ranges,
            { from: currentSelection.from, to: 0, selected: false },
            range => range.from
        )
        const previousRange = idx === 0 ? this.ranges[this.ranges.length - 1] : this.ranges[idx - 1]
        navigateTo(this.view, previousRange)
    }

    private navigateNext() {
        const currentSelection = this.view.state.selection.main
        const idx = sortedIndexBy(
            this.ranges,
            { from: currentSelection.from + 1, to: 0, selected: false },
            range => range.from
        )
        const nextRange = this.ranges[idx % this.ranges.length]
        navigateTo(this.view, nextRange)
    }

    private render(ranges: HighlightedRange[]): void {
        if (!this.root) {
            this.root = createRoot(this.dom)
        }

        const totalMatches = ranges.length
        const selectedIdx = ranges.findIndex(range => range.selected)

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
                    {selectedIdx === undefined ? '' : `${selectedIdx + 1} of `}
                    {totalMatches} {pluralize('result', totalMatches)}
                </Text>
            </CodeMirrorContainer>
        )
    }
}

function navigateTo(view: EditorView, target: { from: number; to: number }) {
    view.dispatch({
        selection: { anchor: target.from, head: target.to },
        effects: [
            EditorView.scrollIntoView(target.from, {
                y: 'nearest',
                x: 'center',
                yMargin: view.dom.getBoundingClientRect().height / 3,
            }),
        ],
    })
}
