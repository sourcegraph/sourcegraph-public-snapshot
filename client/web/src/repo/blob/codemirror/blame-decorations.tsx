/**
 * This file contains CodeMirror extensions for rendering git blame specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link showGitBlameDecorations} facet.
 */
import { Facet, RangeSet } from '@codemirror/state'
import { Decoration, EditorView, gutter, gutterLineClass, GutterMarker } from '@codemirror/view'
import { History } from 'history'
import { isEqual } from 'lodash'
import { createRoot, Root } from 'react-dom/client'

import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

import { BlameHunk } from '../../blame/useBlameHunks'
import { BlameDecoration } from '../BlameDecoration'

import { blobPropsFacet } from '.'

import blameColumnStyles from '../BlameColumn.module.scss'

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

/**
 * Used to find the blame decoration(s) with the longest text,
 * so that they can be used as gutter spacer.
 */
const longestColumnDecorations = (hunks?: BlameHunk[]): BlameHunk | undefined =>
    hunks?.reduce((acc, hunk) => {
        if (!acc || hunk.displayInfo.message.length > acc.displayInfo.message.length) {
            return hunk
        }
        return acc
    }, undefined as BlameHunk | undefined)

/**
 * Widget class for rendering column git blame text document decorations inside CodeMirror.
 */
class BlameDecoratorMarker extends GutterMarker {
    private container: HTMLElement | null = null
    private reactRoot: Root | null = null
    private state: { history: History }

    constructor(
        public view: EditorView,
        public readonly hunk: BlameHunk | undefined,
        private isSpacer: boolean = false
    ) {
        super()
        this.state = { history: this.view.state.facet(blobPropsFacet).history }
    }

    /* eslint-disable-next-line id-length*/
    public eq(other: BlameDecoratorMarker): boolean {
        return isEqual(this.hunk, other.hunk)
    }

    public toDOM(): HTMLElement {
        if (!this.container) {
            this.container = document.createElement('span')
            this.reactRoot = createRoot(this.container)
            this.reactRoot.render(
                <BlameDecoration
                    /* line has to be set to 0 if this marker is used as spacer,
                     * otherwise the popover will be rendered twice when
                     * hovering over the line associated with this hunk
                     */
                    line={this.isSpacer ? 0 : this.hunk?.startLine ?? 0}
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
        gutter({
            class: blameColumnStyles.decoration,
            lineMarker: (view, lineBlock) => {
                const hunks = view.state.facet(facet)
                if (!hunks) {
                    // This shouldn't be possible but just in case
                    return null
                }
                const lineNumber: number = view.state.doc.lineAt(lineBlock.from).number
                const hunk = hunks.find(hunk => hunk.startLine === lineNumber)
                if (!hunk) {
                    return null
                }
                return new BlameDecoratorMarker(view, hunk)
            },
            // Without a spacer the whole gutter flickers when the
            // decorations for the visible lines are re-rendered
            // TODO: update spacer when decorations change
            initialSpacer: view => {
                const hunk = longestColumnDecorations(view.state.facet(facet))
                return new BlameDecoratorMarker(view, hunk, true)
            },
            // Markers need to be updated when theme changes
            lineMarkerChange: update =>
                update.startState.facet(EditorView.darkTheme) !== update.state.facet(EditorView.darkTheme),
        }),
        hoveredLine,
    ],
})
