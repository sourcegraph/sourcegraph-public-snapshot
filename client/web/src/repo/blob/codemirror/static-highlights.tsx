/**
 * This provides CodeMirror extension for highlighting a static set of ranges.
 */
import { type Extension, StateEffect, EditorState, StateField, Facet } from '@codemirror/state'
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
    const facet = Facet.define<Range[], Range[]>({
        combine: ranges => ranges.flat(),
        enables: self => [
            EditorView.decorations.compute([self], state => {
                const ranges = state.facet(self)
                return Decoration.set(
                    ranges.map(range =>
                        staticHighlightDecoration.range(
                            toCodeMirrorLocation(state, range.start),
                            toCodeMirrorLocation(state, range.end)
                        )
                    ),
                    true
                )
            }),
            EditorView.theme({
                '.cm-sg-staticSelection': {
                    backgroundColor: 'var(--mark-bg)',
                },
            }),
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
const navigateSelection = StateEffect.define<number>()

class StaticHighlightsPanel implements Panel {
    public dom: HTMLElement
    public top = true
    private ranges: {from: number, to: number}[] = []

    // Currently selected 0-based match index.
    //private selected: number = 0
    private root: Root | null = null

    constructor(private view: EditorView, private navigate: NavigateFunction, ranges: Range[]) {
        this.ranges = ranges.map(range => ({
            from: toCodeMirrorLocation(view.state, range.start),
            to: toCodeMirrorLocation(view.state, range.end),
        })),
        this.dom = createElement('div', {
            className: classNames('cm-sg-search-container', styles.root),
            id: STATIC_HIGHLIGHTS_CONTAINER_ID,
        })
    }

    public update(update: ViewUpdate): void {
    }

    public mount(): void {
        this.render()
    }

    public destroy(): void { }

    private navigatePrevious() {
        const currentSelection = this.view.state.selection.main
        let previous: {from: number, to: number} | undefined
        for (const range of this.ranges) {
            previous = range
            if (range.from >= currentSelection.from) {
                break
            }
        }
        this.view.dispatch({
            selection: {anchor: previous?.from ?? 0, head: previous?.to ?? 0},
            scrollIntoView: true,
        })
    }

    private navigateNext() {
        const currentSelection = this.view.state.selection.main
        let next: {from: number, to: number} | undefined
        for (const range of this.ranges) {
            next = range
            if (range.from > currentSelection.from) {
                break
            }
        }
        this.view.dispatch({
            selection: {anchor: next?.from ?? 0, head: next?.to ?? 0},
            scrollIntoView: true,
        })
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
