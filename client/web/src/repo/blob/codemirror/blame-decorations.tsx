/**
 * This file contains CodeMirror extensions for rendering git blame specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link showGitBlameDecorations} facet.
 */
import { Facet, RangeSet } from '@codemirror/state'
import {
    Decoration,
    DecorationSet,
    EditorView,
    gutter,
    gutterLineClass,
    GutterMarker,
    gutters,
    ViewPlugin,
    ViewUpdate,
    WidgetType,
} from '@codemirror/view'
import { History } from 'history'
import { isEqual } from 'lodash'
import { createRoot, Root } from 'react-dom/client'

import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

import { BlameHunk } from '../../blame/useBlameHunks'
import { BlameDecoration } from '../BlameDecoration'

import { blobPropsFacet } from '.'

const highlightedLineDecoration = Decoration.line({ class: 'highlighted-line' })
const highlightedLineGutterMarker = new (class extends GutterMarker {
    public elementClass = 'highlighted-line'
})()

const [hoveredLine, setHoveredLine] = createUpdateableField<number | null>(null, field => [
    EditorView.decorations.compute([field], state => {
        const line = state.field(field, false) ?? null
        return line === null
            ? Decoration.none
            : Decoration.set(highlightedLineDecoration.range(state.doc.line(line).from))
    }),
    gutterLineClass.compute([field], state => {
        const line = state.field(field, false) ?? null
        return line === null
            ? RangeSet.empty
            : RangeSet.of(highlightedLineGutterMarker.range(state.doc.line(line).from))
    }),
])

class BlameDecorationWidget extends WidgetType {
    private container: HTMLElement | null = null
    private reactRoot: Root | null = null
    private state: { history: History }

    constructor(public view: EditorView, public readonly hunk: BlameHunk | undefined, public readonly line: number) {
        super()
        this.state = { history: this.view.state.facet(blobPropsFacet).history }
    }

    /* eslint-disable-next-line id-length*/
    public eq(other: BlameDecorationWidget): boolean {
        return isEqual(this.hunk, other.hunk)
    }

    public toDOM(): HTMLElement {
        if (!this.container) {
            this.container = document.createElement('span')
            this.container.classList.add('blame-decoration')

            this.reactRoot = createRoot(this.container)
            this.reactRoot.render(
                <BlameDecoration
                    line={this.hunk?.startLine ?? 0}
                    blameHunk={this.hunk}
                    history={this.state.history}
                    onSelect={this.selectRow}
                    onDeselect={this.deselectRow}
                />
            )
        }
        return this.container
    }

    private selectRow = (line: number): void => {
        setHoveredLine(this.view, line)
    }

    private deselectRow = (line: number): void => {
        if (this.view.state.field(hoveredLine) === line) {
            setHoveredLine(this.view, null)
        }
    }

    public destroy(): void {
        this.container?.remove()
        // setTimeout seems necessary to prevent React from complaining that the
        // root is synchronously unmounted while rendering is in progress
        setTimeout(() => this.reactRoot?.unmount(), 0)
    }
}

/**
 * Facet to show git blame decorations.
 */
export const showGitBlameDecorations = Facet.define<BlameHunk[], BlameHunk[]>({
    combine: decorations => decorations.flat(),
    enables: facet => [
        hoveredLine,

        // Render blame hunks as line decorations.
        ViewPlugin.fromClass(
            class {
                public decorations: DecorationSet
                private previousHunkLength = -1

                constructor(view: EditorView) {
                    this.decorations = this.computeDecorations(view, facet)
                }

                public update(update: ViewUpdate): void {
                    const hunks = update.view.state.facet(facet)

                    if (update.docChanged || update.viewportChanged || this.previousHunkLength !== hunks.length) {
                        this.decorations = this.computeDecorations(update.view, facet)
                        this.previousHunkLength = update.state.facet(facet).length
                    }
                }

                private computeDecorations(view: EditorView, facet: Facet<BlameHunk[], BlameHunk[]>): DecorationSet {
                    const widgets = []
                    const hunks = view.state.facet(facet)
                    for (const { from, to } of view.visibleRanges) {
                        for (let position = from; position <= to; ) {
                            const line = view.state.doc.lineAt(position)
                            const matchingHunk = hunks.find(hunk => hunk.startLine === line.number)
                            const decoration = Decoration.widget({
                                widget: new BlameDecorationWidget(view, matchingHunk, line.number),
                            })
                            widgets.push(decoration.range(line.from))
                            position = line.to + 1
                        }
                    }
                    return Decoration.set(widgets)
                }
            },
            {
                decorations: ({ decorations }) => decorations,
            }
        ),
        EditorView.theme({
            '.cm-line': {
                // Position relative so that the blame-decoration inside can be
                // aligned to the start of the line
                position: 'relative',
                // Move the start of the line to after the blame decoration.
                // This is necessary because the start of the line is used for
                // aligning tab characters.
                //
                // 1rem is the default padding-left so we have to add it here
                paddingLeft: 'calc(var(--blame-decoration-width) + 1rem) !important',
            },

            '.blame-decoration': {
                // Remove the blame decoration from the content flow so that
                // the tab start can be moved to the right
                position: 'absolute',
                left: '0',

                display: 'inline-block',
                background: 'var(--body-bg)',
                verticalAlign: 'bottom',
                width: 'var(--blame-decoration-width)',
            },

            '.selected-line .blame-decoration, .highlighted-line .blame-decoration': {
                background: 'inherit',
            },

            '.cm-content': {
                // Make .cm-content overflow .blame-gutter
                marginLeft: 'calc(var(--blame-decoration-width) * -1)',
                // override default .cm-gutters z-index 200
                zIndex: 201,
            },
        }),
    ],
})

const blameGutterElement = new (class extends GutterMarker {
    public toDOM(): HTMLElement {
        return document.createElement('div')
    }
})()

export const showBlameGutter = Facet.define<boolean>({
    combine: value => value.flat(),
    enables: [
        // Render gutter with no content only to create a column with specified background.
        // This column is used by .cm-content shifted to the left by var(--blame-decoration-width)
        // to achieve column-like view of inline blame decorations.
        gutter({
            class: 'blame-gutter',
            lineMarker: () => blameGutterElement,
            initialSpacer: () => blameGutterElement,
        }),

        // By default, gutters are fixed, meaning they don't scroll along with the content horizontally (position: sticky).
        // We override this behavior when blame decorations are shown to make inline decorations column-like view work.
        gutters({ fixed: false }),

        EditorView.theme({
            '.blame-gutter': {
                background: 'var(--body-bg)',
                width: 'var(--blame-decoration-width)',
            },
        }),
    ],
})
