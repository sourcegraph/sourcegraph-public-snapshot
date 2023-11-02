/**
 * This file contains CodeMirror extensions for rendering git blame specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link blameData} facet.
 */
import { type Extension, Facet, RangeSetBuilder } from '@codemirror/state'
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
    type PluginValue,
} from '@codemirror/view'
import { isEqual } from 'lodash'

import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

import type { BlameHunk, BlameHunkData } from '../../blame/useBlameHunks'
import { getBlameRecencyColor } from '../blameRecency'

/**
 * Unlike {@link BlameHunkData} which is a list of unordered hunks, this
 * structure provides a line number -> blame hunk index for fast access.
 */
interface IndexedBlameHunkData extends Pick<BlameHunkData, 'firstCommitDate' | 'externalURLs'> {
    lines: BlameHunk[]
}

const highlightedLineDecoration = Decoration.line({ class: 'highlighted-line' })
const startOfHunkDecoration = Decoration.line({ class: 'border-top' })

const highlightedLineGutterMarker = new (class extends GutterMarker {
    public elementClass = 'highlighted-line'
})()

/**
 * Used for styling the currently hovered hunk.
 */
const [hoveredHunk, setHoveredHunk] = createUpdateableField<BlameHunk | null>(null, field => [
    EditorView.decorations.compute([field], state => {
        const hunk = state.field(field)
        const builder = new RangeSetBuilder<Decoration>()
        if (hunk) {
            for (let line = hunk.startLine; line < hunk.endLine; line++) {
                const from = state.doc.line(line).from
                builder.add(from, from, highlightedLineDecoration)
            }
        }
        return builder.finish()
    }),
    gutterLineClass.compute([field], state => {
        const hunk = state.field(field)
        const builder = new RangeSetBuilder<GutterMarker>()
        if (hunk) {
            for (let line = hunk.startLine; line < hunk.endLine; line++) {
                const from = state.doc.line(line).from
                builder.add(from, from, highlightedLineGutterMarker)
            }
        }
        return builder.finish()
    }),
])

class BlameDecorationWidget extends WidgetType {
    private container: HTMLElement | null = null
    private decoration: ReturnType<BlameConfig['createBlameDecoration']> | undefined

    constructor(
        public view: EditorView,
        public readonly hunk: BlameHunk,
        public readonly line: number,
        public readonly externalURLs: BlameHunkData['externalURLs'],
        private readonly createBlameDecoration: BlameConfig['createBlameDecoration']
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

            this.decoration = this.createBlameDecoration(this.container, {
                line: this.line,
                hunk: this.hunk,
                onSelect: this.select,
                onDeselect: this.deselect,
                externalURLs: this.externalURLs,
            })
        }
        return this.container
    }

    private select = (): void => {
        setHoveredHunk(this.view, this.hunk)
    }

    private deselect = (): void => {
        if (this.view.state.field(hoveredHunk, false) === this.hunk) {
            setHoveredHunk(this.view, null)
        }
    }

    public destroy(): void {
        this.container?.remove()
        // setTimeout seems necessary to prevent React from complaining that the
        // root is synchronously unmounted while rendering is in progress
        setTimeout(() => this.decoration?.destroy?.(), 0)
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
        lineHeight: '1.5rem',
        // Avoid jumping when blame decorations are streamed in because we use a border
        // borderTop: '1px solid transparent',
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

    constructor(private view: EditorView, private config: BlameConfig) {
        this.decorations = this.computeDecorations(view.state.facet(blameDataFacet))
    }

    public update(update: ViewUpdate): void {
        const blame = update.state.facet(blameDataFacet)

        if (update.docChanged || update.viewportChanged || blame !== update.startState.facet(blameDataFacet)) {
            this.decorations = this.computeDecorations(blame)
        }
    }

    private computeDecorations({ lines, externalURLs }: IndexedBlameHunkData): DecorationSet {
        const view = this.view
        const builder = new RangeSetBuilder<Decoration>()

        // Keeps track of the last found hunk. We don't want to show an additional blame
        // decoration when a range inside a single hunk is shown.
        let previousHunk: BlameHunk | undefined

        for (const { from, to } of view.visibleRanges) {
            // We add + 1 because it appears that the range after a folded region
            // starts at the previous line (maybe because of line break?). So to
            // correctly show hunk information that starts in the folded range we
            // have to start at the line after the fold.
            let line = view.state.doc.lineAt(from + 1)
            const endLine = view.state.doc.lineAt(to).number
            while (line.number <= endLine) {
                const matchingHunk = lines[line.number]
                if (matchingHunk && matchingHunk.rev !== previousHunk?.rev) {
                    if (line.number !== 1) {
                        builder.add(line.from, line.from, startOfHunkDecoration)
                    }
                    builder.add(
                        line.from,
                        line.from,
                        Decoration.widget({
                            widget: new BlameDecorationWidget(
                                view,
                                matchingHunk,
                                line.number,
                                externalURLs,
                                this.config.createBlameDecoration
                            ),
                        })
                    )
                    previousHunk = matchingHunk
                }
                try {
                    // endLine is exclusive
                    // this throws when endLine is not a valid line in the document
                    line = view.state.doc.line(matchingHunk?.endLine ?? line.number + 1)
                } catch {
                    break
                }
            }
        }
        return builder.finish()
    }
}

/**
 * Facet to show git blame decorations.
 */
const blameDataFacet = Facet.define<BlameHunkData, IndexedBlameHunkData>({
    combine(values) {
        const value = values[0] ?? { current: [] }
        const lines = []
        for (const hunk of value.current ?? []) {
            for (let i = hunk.startLine; i < hunk.endLine; i++) {
                lines[i] = hunk
            }
        }
        return {
            lines,
            firstCommitDate: value.firstCommitDate,
            externalURLs: value.externalURLs,
        }
    },
})

class RecencyMarker extends GutterMarker {
    // hunk can be undefined if when the data is not available yet
    constructor(private line: number, private hunk?: BlameHunk) {
        super()
    }

    public eq(other: RecencyMarker): boolean {
        // Only consider two markers with the same line equal if
        // hunk data is available. Otherwise the marker won't be
        // update/recreated as new data becomes available.
        return this.line === other.line && !!this.hunk && !!other.hunk
    }

    public toDOM(view: EditorView): Node {
        const dom = document.createElement('div')
        dom.className = 'sg-recency-marker'
        const { firstCommitDate } = view.state.facet(blameDataFacet)
        if (this.hunk) {
            if (this.hunk.startLine === this.line) {
                dom.classList.add('border-top')
            }
            dom.style.backgroundColor = getBlameRecencyColor(new Date(this.hunk.author.date), firstCommitDate)
        }
        return dom
    }
}

const blameGutter: Extension = [
    // By default, gutters are fixed, meaning they don't scroll along with the content horizontally (position: sticky).
    // We override this behavior when blame decorations are shown to make inline decorations column-like view work.
    gutters({ fixed: false }),

    // Gutter for recency indicator
    gutter({
        lineMarker(view, line) {
            const lineNumber = view.state.doc.lineAt(line.from).number
            const hunks = view.state.facet(blameDataFacet).lines
            return new RecencyMarker(lineNumber, hunks[lineNumber])
        },
        lineMarkerChange(update) {
            return update.state.facet(blameDataFacet) !== update.startState.facet(blameDataFacet)
        },
    }),

    // Render gutter with no content only to create a column with specified background.
    // This column is used by .cm-content shifted to the left by var(--blame-decoration-width)
    // to achieve column-like view of inline blame decorations.
    gutter({
        class: 'blame-gutter',
    }),

    EditorView.theme({
        '.blame-gutter': {
            background: 'var(--body-bg)',
            width: 'var(--blame-decoration-width)',
        },
        '.sg-recency-marker': {
            position: 'relative',
            height: '100%',
            width: 'var(--blame-recency-width)',
        },
    }),
]

interface BlameConfig {
    createBlameDecoration: (
        container: HTMLElement,
        spec: {
            line: number
            hunk: BlameHunk
            onSelect: () => void
            onDeselect: () => void
            externalURLs: BlameHunkData['externalURLs']
        }
    ) => { destroy?: () => void }
}

/**
 * Show blame column.
 */
export function showBlame(config: BlameConfig): Extension {
    return [
        blameGutter,
        hoveredHunk,
        blameDecorationTheme,
        ViewPlugin.define(view => new BlameDecorationViewPlugin(view, config), {
            decorations: ({ decorations }) => decorations,
        }),
    ]
}

/**
 * Provide blame data.
 */
export function blameData(blameHunks: BlameHunkData | undefined): Extension {
    return blameHunks ? blameDataFacet.of(blameHunks) : []
}
