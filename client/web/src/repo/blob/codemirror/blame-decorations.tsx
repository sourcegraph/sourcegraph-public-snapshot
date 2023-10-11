/**
 * This file contains CodeMirror extensions for rendering git blame specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link showGitBlameDecorations} facet.
 */
import { type Extension, Facet, RangeSet } from '@codemirror/state'
import {
    Decoration,
    type DecorationSet,
    EditorView,
    gutter,
    gutterLineClass,
    GutterMarker,
    gutters,
    ViewPlugin,
    type ViewUpdate,
    WidgetType,
    PluginValue,
} from '@codemirror/view'
import { isEqual } from 'lodash'
import { createRoot, type Root } from 'react-dom/client'

import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

import type { BlameHunk, BlameHunkData } from '../../blame/useBlameHunks'
import { BlameDecoration } from '../BlameDecoration'
import { NavigateFunction } from 'react-router-dom'

const highlightedLineDecoration = Decoration.line({ class: 'highlighted-line' })
const startOfHunkDecoration = Decoration.line({ class: 'border-top' })

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

    constructor(
        public view: EditorView,
        public readonly hunk: BlameHunk | undefined,
        public readonly line: number,
        public readonly blameHunkMetadata: Omit<BlameHunkData, 'current'>,
        private readonly navigate: NavigateFunction
    ) {
        super()
    }

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
                    navigate={this.navigate}
                    line={this.line ?? 0}
                    blameHunk={this.hunk}
                    onSelect={this.selectRow}
                    onDeselect={this.deselectRow}
                    firstCommitDate={this.blameHunkMetadata.firstCommitDate}
                    externalURLs={this.blameHunkMetadata.externalURLs}
                    hideRecency={false}
                />
            )
        }
        return this.container
    }

    private selectRow = (line: number): void => {
        setHoveredLine(this.view, line)
    }

    private deselectRow = (line: number): void => {
        if (this.view.state.field(hoveredLine, false) === line) {
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

const blameDecorationTheme = EditorView.theme({
    '.cm-line': {
        // Position relative so that the blame-decoration inside can be
        // aligned to the start of the line
        position: 'relative',
        // Move the start of the line to after the blame decoration.
        // This is necessary because the start of the line is used for
        // aligning tab characters.
        paddingLeft: 'var(--blame-decoration-width) !important',
    },
    '.blame-decoration': {
        // Remove the blame decoration from the content flow so that
        // the tab start can be moved to the right
        position: 'absolute',
        left: '0',
        height: '100%',
        display: 'inline-block',
        background: 'var(--body-bg)',
        verticalAlign: 'bottom',
        width: 'var(--blame-decoration-width)',

        '.selected-line &, .highlighted-line &': {
            background: 'inherit',
        },
    },

    '.cm-content': {
        // Make .cm-content overflow .blame-gutter
        marginLeft: 'calc(var(--blame-decoration-width) * -1)',
        // override default .cm-gutters z-index 200
        zIndex: 201,
    },
})

class BlameDecorationViewPlugin implements PluginValue {
    public decorations: DecorationSet
    private previousHunkLength = -1
    private previousBlameHunkMetadata: Omit<BlameHunkData, 'current'> | undefined

    constructor(private view: EditorView, private navigate: NavigateFunction) {
        this.decorations = this.computeDecorations(view.state.facet(showGitBlameDecorations))
    }

    public update(update: ViewUpdate): void {
        const facetProps = update.view.state.facet(showGitBlameDecorations)
        const hunks = facetProps.hunks
        const blameHunkMetadata = facetProps.blameHunkMetadata

        if (
            update.docChanged ||
                update.viewportChanged ||
                this.previousHunkLength !== hunks.length ||
                this.previousBlameHunkMetadata !== blameHunkMetadata
        ) {
            this.decorations = this.computeDecorations(facetProps)
            this.previousHunkLength = hunks.length
            this.previousBlameHunkMetadata = blameHunkMetadata
        }
    }

    private computeDecorations({hunks, blameHunkMetadata}: BlameDecorationsFacetProps): DecorationSet {
        const view = this.view
        const widgets = []

        for (const { from, to } of view.visibleRanges) {
            let nextHunkDecorationLineRenderedAt = -1
            for (let position = from; position <= to; ) {
                const line = view.state.doc.lineAt(position)
                const matchingHunk = hunks.find(
                    hunk => line.number >= hunk.startLine && line.number < hunk.endLine
                )

                const isStartOfHunk = matchingHunk?.startLine === line.number
                if (
                    (isStartOfHunk && line.number !== 1) ||
                        nextHunkDecorationLineRenderedAt === line.from
                ) {
                    widgets.push(startOfHunkDecoration.range(line.from))

                    // When we found a hunk, we already know when the next one will start even if this
                    // hunk was not loaded yet.
                    //
                    // We mark this as rendered in `nextHunkDecorationLineRenderedAt` so that the next
                    // startLine can be skipped if it was rendered already
                    if (matchingHunk) {
                        nextHunkDecorationLineRenderedAt = view.state.doc.line(matchingHunk.endLine).from
                    }
                }

                const decoration = Decoration.widget({
                    widget: new BlameDecorationWidget(
                        view,
                        matchingHunk,
                        line.number,
                        blameHunkMetadata,
                        this.navigate,
                    ),
                })
                widgets.push(decoration.range(line.from))
                position = line.to + 1
            }
        }
        return Decoration.set(widgets)
    }
}

/**
 * Facet to show git blame decorations.
 */
interface BlameDecorationsFacetProps {
    hunks: BlameHunk[]
    blameHunkMetadata: Omit<BlameHunkData, 'current'>
}
const showGitBlameDecorations = Facet.define<BlameDecorationsFacetProps, BlameDecorationsFacetProps>({
    combine: decorations => decorations[0] ?? {hunks: [], blameHunkMetadata: {}},
})

const blameGutterElement = new (class extends GutterMarker {
    public toDOM(): HTMLElement {
        return document.createElement('div')
    }
})()

const blameGutter: Extension = [
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
    })
]

const blameLineTheme = EditorView.theme({
    '.cm-line': {
        lineHeight: '1.5rem',
        // Avoid jumping when blame decorations are streamed in because we use a border
        borderTop: '1px solid transparent',
    },
})

export function showBlame(navigate: NavigateFunction): Extension {
    return [
        EditorView.editorAttributes.of({ class: 'sg-blame-visible'}),
        blameLineTheme,
        blameGutter,
        hoveredLine,
        blameDecorationTheme,
        ViewPlugin.define(
            view => new BlameDecorationViewPlugin(view, navigate),
            {
                decorations: ({ decorations }) => decorations,
            }
        ),
    ]
}

export function blameData(blameHunks: BlameHunkData | undefined): Extension {
    return blameHunks?.current
        ? showGitBlameDecorations.of({
              hunks: blameHunks.current,
              blameHunkMetadata: {
                  firstCommitDate: blameHunks.firstCommitDate,
                  externalURLs: blameHunks.externalURLs,
              },
          }): []

}
