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
    gutterLineClass,
    GutterMarker,
    ViewPlugin,
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

class DecorationWidget extends WidgetType {
    private container: HTMLElement | null = null
    private reactRoot: Root | null = null
    private state: { history: History }

    constructor(public view: EditorView, public readonly hunk: BlameHunk | undefined) {
        super()
        this.state = { history: this.view.state.facet(blobPropsFacet).history }
    }

    /* eslint-disable-next-line id-length*/
    public eq(other: DecorationWidget): boolean {
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

const decorate = (view: EditorView, facet: Facet<BlameHunk[], BlameHunk[]>): DecorationSet => {
    const widgets = []
    const hunks = view.state.facet(facet)
    for (const { from, to } of view.visibleRanges) {
        for (let position = from; position <= to; ) {
            const line = view.state.doc.lineAt(position)
            const hunk = hunks.find(h => h.startLine === line.number)
            const decoration = Decoration.widget({
                widget: new DecorationWidget(view, hunk),
            })
            widgets.push(decoration.range(line.from))
            position = line.to + 1
        }
    }
    return Decoration.set(widgets)
}

/**
 * Facet to show git blame decorations.
 */
export const showGitBlameDecorations = Facet.define<BlameHunk[], BlameHunk[]>({
    combine: decorations => decorations.flat(),
    enables: facet => [
        hoveredLine,
        ViewPlugin.fromClass(
            class {
                public decorations: DecorationSet

                constructor(view: EditorView) {
                    this.decorations = decorate(view, facet)
                }
            },
            {
                decorations: ({ decorations }) => decorations,
            }
        ),
        EditorView.theme({
            '.cm-line': {
                paddingLeft: '0 !important', // line code content left padding is handled by .blame-decoration right margin
            },

            '.blame-decoration': {
                display: 'inline-block',
                background: 'var(--body-bg)',
                verticalAlign: 'bottom',

                marginRight: '1rem',
            },

            '.selected-line .blame-decoration, .highlighted-line .blame-decoration': {
                background: 'inherit',
            },
        }),
    ],
})
