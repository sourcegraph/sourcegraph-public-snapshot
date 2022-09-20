/**
 * This file contains CodeMirror extensions for rendering git blame specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link showGitBlameDecorations} facet.
 */
import { Compartment, Extension, Facet } from '@codemirror/state'
import {
    Decoration,
    DecorationSet,
    EditorView,
    gutter,
    GutterMarker,
    PluginValue,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'
import { isEqual } from 'lodash'
import { createRoot, Root } from 'react-dom/client'

import { BlameHunk } from '../../blame/useBlameHunks'
import { BlameDecoration } from '../BlameDecoration'

import blameColumnStyles from '../BlameColumn.module.scss'

/**
 * {@link BlameDecorationManager} creates {@link gutter} dynamically.
 * Using a compartment allows us to change the gutter extensions without
 * impacting any other extensions.
 */
const decorationGutters = new Compartment()

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
 * Get HTMLElement containing line number by line number.
 * */
const getLineNumberCell = (line: number): HTMLElement | null =>
    document.querySelector<HTMLElement>(`.cm-editor .cm-gutters .cm-gutterElement:nth-of-type(${line + 1})`)

/**
 * Get HTMLElement containing code chunk by line number.
 * */
const getCodeCell = (line: number): HTMLElement | null =>
    document.querySelector<HTMLElement>(`.cm-editor .cm-content .cm-line:nth-of-type(${line})`)

/**
 * Widget class for rendering column git blame text document decorations inside CodeMirror.
 */
class BlameDecoratorMarker extends GutterMarker {
    private container: HTMLElement | null = null
    private reactRoot: Root | null = null

    constructor(public readonly item: BlameHunk | undefined, public readonly line: number) {
        super()
    }

    /* eslint-disable-next-line id-length*/
    public eq(other: BlameDecoratorMarker): boolean {
        return isEqual(this.item, other.item)
    }

    public toDOM(): HTMLElement {
        if (!this.container) {
            this.container = document.createElement('span')
            this.reactRoot = createRoot(this.container)
            this.reactRoot.render(
                <BlameDecoration
                    line={this.line}
                    blameHunk={this.item}
                    onSelect={this.selectRow}
                    onDeselect={this.deselectRow}
                />
            )
        }
        return this.container
    }

    private getDecorationCell = (): HTMLElement | null | undefined => this.container?.closest('.cm-gutterElement')

    private selectRow = (line: number): void => {
        const lineNumberCell = getLineNumberCell(line)
        const decorationCell = this.getDecorationCell()
        const codeCell = getCodeCell(line)
        if (!lineNumberCell || !decorationCell || !codeCell) {
            return
        }

        for (const cell of [lineNumberCell, decorationCell, codeCell]) {
            cell.classList.add('highlighted-line')
        }
    }

    private deselectRow = (line: number): void => {
        const lineNumberCell = getLineNumberCell(line)
        const decorationCell = this.getDecorationCell()
        const codeCell = getCodeCell(line)
        if (!lineNumberCell || !decorationCell || !codeCell) {
            return
        }

        for (const cell of [lineNumberCell, decorationCell, codeCell]) {
            cell.classList.remove('highlighted-line')
        }
    }

    public destroy(): void {
        this.container?.remove()
        // setTimeout seems necessary to prevent React from complaining that the
        // root is synchronously unmounted while rendering is in progress
        setTimeout(() => this.reactRoot?.unmount(), 0)
    }
}

class BlameDecorationManager implements PluginValue {
    public decorations: DecorationSet = Decoration.none
    private gutter: { gutter: Extension; items: BlameHunk[] } | undefined
    private reset: number | null = null

    constructor(private readonly view: EditorView) {
        this.updateDecorations(view.state.facet(showGitBlameDecorations))
    }

    public update(update: ViewUpdate): void {
        const currentDecorations = update.state.facet(showGitBlameDecorations)

        if (update.startState.facet(showGitBlameDecorations) !== currentDecorations) {
            this.updateDecorations(currentDecorations)
        }
    }

    private updateDecorations(specs: BlameHunk[]): void {
        if (this.updateGutter(specs)) {
            // We cannot synchronously dispatch another transaction during
            // an update, so we schedule it but also cancel pending
            // transactions should this be called multiple times in a row
            if (this.reset !== null) {
                window.clearTimeout(this.reset)
            }
            this.reset = window.setTimeout(() => {
                this.view.dispatch({
                    effects: decorationGutters.reconfigure(this.gutter?.gutter || []),
                })
            }, 50)
        }
    }

    /**
     * Create or remove gutters.
     */
    private updateGutter(specs: BlameHunk[]): boolean {
        let change = false

        if (!this.gutter) {
            this.gutter = {
                gutter: gutter({
                    class: blameColumnStyles.decoration,
                    lineMarker: (view, lineBlock) => {
                        const items = this.gutter?.items
                        if (!items) {
                            // This shouldn't be possible but just in case
                            return null
                        }
                        const lineNumber: number = view.state.doc.lineAt(lineBlock.from).number
                        const lineItems = items.find(hunk => hunk.startLine === lineNumber)
                        if (!lineItems) {
                            return null
                        }
                        return new BlameDecoratorMarker(lineItems, lineNumber)
                    },
                    // Without a spacer the whole gutter flickers when the
                    // decorations for the visible lines are re-rendered
                    // TODO: update spacer when decorations change
                    initialSpacer: () => {
                        const hunk = longestColumnDecorations(this.gutter?.items)
                        return new BlameDecoratorMarker(hunk, 0)
                    },
                    // Markers need to be updated when theme changes
                    lineMarkerChange: update =>
                        update.startState.facet(EditorView.darkTheme) !== update.state.facet(EditorView.darkTheme),
                }),
                items: specs,
            }
            change = true
        } else {
            this.gutter.items = specs
        }

        return change
    }
}

/**
 * Facet to show git blame decorations.
 */
export const showGitBlameDecorations = Facet.define<BlameHunk[], BlameHunk[]>({
    combine: decorations => decorations.flat(),
    compareInput: (a, b) => a === b || (a.length === 0 && b.length === 0),
    enables: [
        ViewPlugin.fromClass(BlameDecorationManager, {
            decorations: manager => manager.decorations,
        }),
        decorationGutters.of([]),
    ],
})
