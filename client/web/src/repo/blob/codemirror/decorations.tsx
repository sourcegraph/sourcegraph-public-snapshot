/**
 * This file contains CodeMirror extensions for rendering Sourcegraph specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link showTextDocumentDecorations} facet.
 */
import { Compartment, Extension, Facet, RangeSetBuilder } from '@codemirror/state'
import {
    Decoration,
    DecorationSet,
    EditorView,
    gutter,
    GutterMarker,
    PluginValue,
    ViewPlugin,
    ViewUpdate,
    WidgetType,
} from '@codemirror/view'
import { isEqual } from 'lodash'
import { createRoot, Root } from 'react-dom/client'
import { TextDocumentDecorationType } from 'sourcegraph'

import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { DecorationMapByLine, groupDecorationsByLine } from '@sourcegraph/shared/src/api/extension/api/decorations'

import { BlameHunk } from '../../blame/useBlameDecorations'
import { ColumnDecoratorContents } from '../ColumnDecorator'
import { LineDecoratorContents } from '../LineDecorator'

import columnDecoratorStyles from '../ColumnDecorator.module.scss'
import lineDecoratorStyles from '../LineDecorator.module.scss'

export type TextDocumentDecorationSpec = [TextDocumentDecorationType, TextDocumentDecoration[]]

/**
 * {@link TextDocumentDecorationMangar} creates {@link gutter}s dynamically.
 * Using a compartment allows us to change the gutter extensions without
 * impacting any other extensions.
 */
const decorationGutters = new Compartment()

/**
 * This class manages inline and column text document decorations. Column
 * decorations are rendered via (dynamically created) gutters, inline
 * decorations are rendered as widget decorations.
 */
class TextDocumentDecorationManager implements PluginValue {
    public inlineDecorations: DecorationSet = Decoration.none
    public decorations: DecorationMapByLine = new Map()

    constructor(private readonly view: EditorView) {
        this.updateDecorations(view.state.facet(showTextDocumentDecorations), !view.state.facet(EditorView.darkTheme))
    }

    public update(update: ViewUpdate): void {
        const currentDecorations = update.state.facet(showTextDocumentDecorations)
        const isLightTheme = !update.state.facet(EditorView.darkTheme)

        if (update.startState.facet(showTextDocumentDecorations) !== currentDecorations) {
            this.updateDecorations(currentDecorations, isLightTheme)
        } else if (update.viewportChanged || isLightTheme !== !update.startState.facet(EditorView.darkTheme)) {
            this.updateInlineDecorations(this.decorations, isLightTheme)
        }
    }

    private updateDecorations(specs: TextDocumentDecorationSpec[], isLightTheme: boolean): void {
        this.decorations = groupDecorationsByLine(
            specs.reduce((acc, [, items]) => [...acc, ...items], [] as TextDocumentDecoration[])
        )
        this.updateInlineDecorations(this.decorations, isLightTheme)
    }

    private updateInlineDecorations(decorations: DecorationMapByLine, isLightTheme: boolean): void {
        const builder = new RangeSetBuilder<Decoration>()

        // Only render decorations for the currently visible lines
        for (const lineBlock of this.view.viewportLineBlocks) {
            const line = this.view.state.doc.lineAt(lineBlock.from).number
            const lineDecorations = decorations.get(line)
            if (lineDecorations) {
                builder.add(
                    lineBlock.to,
                    lineBlock.to,
                    Decoration.widget({ widget: new LineDecorationWidget(lineDecorations, isLightTheme), side: 1 })
                )
            }
        }
        this.inlineDecorations = builder.finish()
    }
}

/**
 * Widget class for rendering inline Sourcegrpah text document decorations inside
 * CodeMirror.
 */
class LineDecorationWidget extends WidgetType {
    private container: HTMLElement | null = null
    private reactRoot: Root | null = null

    constructor(public readonly decorations: TextDocumentDecoration[], public readonly isLightTheme: boolean) {
        super()
    }

    /* eslint-disable-next-line id-length*/
    public eq(other: LineDecorationWidget): boolean {
        // This avoids re-rendering (and flickering) of decorations when the
        // source emits the same decoration.
        return isEqual(this.decorations, other.decorations) && this.isLightTheme === other.isLightTheme
    }

    public toDOM(): HTMLElement {
        if (!this.container) {
            this.container = document.createElement('span')
            this.reactRoot = createRoot(this.container)
            this.reactRoot.render(
                <LineDecoratorContents decorations={this.decorations} isLightTheme={this.isLightTheme} />
            )
        }
        return this.container
    }

    public updateDOM(): boolean {
        if (this.reactRoot) {
            this.reactRoot.render(
                <LineDecoratorContents decorations={this.decorations} isLightTheme={this.isLightTheme} />
            )
            return true
        }
        return false
    }

    public destroy(): void {
        this.container?.remove()
        // setTimeout seems necessary to prevent React from complaining that the
        // root is synchronously unmounted while rendering is in progress
        setTimeout(() => this.reactRoot?.unmount(), 0)
    }
}

/**
 * Used to find the decoration(s) with the longest text, so that they can be
 * used as gutter spacer.
 */
function longestColumnDecorations(hunks: BlameHunk[] | undefined): BlameHunk | undefined {
    return hunks?.reduce((acc, hunk) => {
        if (!acc || hunk.displayInfo.message.length > acc.displayInfo.message.length) {
            return hunk
        }
        return acc
    }, undefined as BlameHunk | undefined)
}

const getLineNumberCell = (line: number): HTMLElement | null =>
    document.querySelector<HTMLElement>(`.cm-editor .cm-gutters .cm-gutterElement:nth-of-type(${line + 1})`)
const getCodeCell = (line: number): HTMLElement | null =>
    document.querySelector<HTMLElement>(`.cm-editor .cm-content .cm-line:nth-of-type(${line})`)

/**
 * Widget class for rendering column Sourcegrpah text document decorations inside
 * CodeMirror.
 */
class ColumnDecoratorMarker extends GutterMarker {
    private container: HTMLElement | null = null
    private reactRoot: Root | null = null

    constructor(
        public readonly item: BlameHunk | undefined,
        public readonly isLightTheme: boolean,
        public readonly line: number
    ) {
        super()
    }

    /* eslint-disable-next-line id-length*/
    public eq(other: ColumnDecoratorMarker): boolean {
        // TODO: Find out why gutter markers still flicker when gutter is
        // redrawn
        return isEqual(this.item, other.item)
    }

    public toDOM(): HTMLElement {
        if (!this.container) {
            this.container = document.createElement('span')
            this.reactRoot = createRoot(this.container)
            this.reactRoot.render(
                <ColumnDecoratorContents
                    line={this.line}
                    blameHunk={this.item}
                    isLightTheme={this.isLightTheme}
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

// Needed to make sure column and inline decorations are rendered correctly
const columnTheme = EditorView.theme({
    [`.${columnDecoratorStyles.decoration} a`]: {
        width: '100%',
    },
    [`.${lineDecoratorStyles.contents}::before`]: {
        content: 'attr(data-contents)',
    },
    [`.${columnDecoratorStyles.decoration}`]: {
        padding: '0',
    },
    [`.${columnDecoratorStyles.item}`]: {
        padding: '0 0.75rem',
    },
})

/**
 * Facet to allow extensions to provide Sourcegraph text document decorations.
 */
export const showTextDocumentDecorations = Facet.define<TextDocumentDecorationSpec[], TextDocumentDecorationSpec[]>({
    combine: decorations => decorations.flat(),
    compareInput: (a, b) => a === b || (a.length === 0 && b.length === 0),
    enables: [
        ViewPlugin.fromClass(TextDocumentDecorationManager, {
            decorations: manager => manager.inlineDecorations,
        }),
    ],
})

class GitBlameDecorationManager implements PluginValue {
    public inlineDecorations: DecorationSet = Decoration.none
    private gutters: Map<string, { gutter: Extension; items: BlameHunk[] }> = new Map()
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
        if (this.updateGutters(specs)) {
            // We cannot synchronously dispatch another transaction during
            // an update, so we schedule it but also cancel pending
            // transactions should this be called multiple times in a row
            if (this.reset !== null) {
                window.clearTimeout(this.reset)
            }
            this.reset = window.setTimeout(() => {
                this.view.dispatch({
                    effects: decorationGutters.reconfigure(Array.from(this.gutters.values(), ({ gutter }) => gutter)),
                })
            }, 50)
        }
    }

    /**
     * Create or remove gutters.
     */
    private updateGutters(specs: BlameHunk[]): boolean {
        let change = false
        const seen: Set<string> = new Set()

        seen.add('blame')
        if (!this.gutters.has('blame')) {
            this.gutters.set('blame', {
                gutter: gutter({
                    class: columnDecoratorStyles.decoration,
                    lineMarker: (view, lineBlock) => {
                        const items = this.gutters.get('blame')?.items
                        if (!items) {
                            // This shouldn't be possible but just in case
                            return null
                        }
                        const lineNumber: number = view.state.doc.lineAt(lineBlock.from).number
                        const lineItems = items.find(hunk => hunk.startLine === lineNumber)
                        if (!lineItems) {
                            return null
                        }
                        return new ColumnDecoratorMarker(lineItems, !view.state.facet(EditorView.darkTheme), lineNumber)
                    },
                    // Without a spacer the whole gutter flickers when the
                    // decorations for the visible lines are re-rendered
                    // TODO: update spacer when decorations change
                    initialSpacer: () => {
                        const hunk = longestColumnDecorations(this.gutters.get('blame')?.items)
                        return new ColumnDecoratorMarker(hunk, /* value doesn't matter for spacer */ true, 0)
                    },
                    // Markers need to be updated when theme changes
                    lineMarkerChange: update =>
                        update.startState.facet(EditorView.darkTheme) !== update.state.facet(EditorView.darkTheme),
                }),
                items: specs,
            })
            change = true
        } else {
            this.gutters.get('blame')!.items = specs
        }
        // }

        for (const id of this.gutters.keys()) {
            if (!seen.has(id)) {
                this.gutters.delete(id)
                change = true
            }
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
        ViewPlugin.fromClass(GitBlameDecorationManager, {
            decorations: manager => manager.inlineDecorations,
        }),
        decorationGutters.of([]),
        columnTheme,
    ],
})
