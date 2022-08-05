/**
 * This file contains CodeMirror extensions for rendering Sourcegraph specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link showTextDocumentDecorations} facet.
 */
import { Compartment, Extension, Facet, Prec, RangeSetBuilder } from '@codemirror/state'
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
import { DecorationMapByLine } from '@sourcegraph/shared/src/api/extension/api/decorations'

import { groupDecorations } from '../Blob'
import { ColumnDecoratorContents } from '../ColumnDecorator'
import { LineDecoratorContents } from '../LineDecorator'

import columnDecoratorStyles from '../ColumnDecorator.module.scss'
import lineDecoratorStyles from '../LineDecorator.module.scss'

export type TextDocumentDecorationSpec = [TextDocumentDecorationType, TextDocumentDecoration[]]
type GroupedDecorations = ReturnType<typeof groupDecorations>

/**
 * Facet to specify whether or not enable column decorations.
 */
export const enableExtensionsDecorationsColumnView = Facet.define<boolean, boolean>({
    combine: values => values[0],
    static: true,
})

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
    public groupedDecorations: GroupedDecorations = { column: [], inline: new Map() }
    private gutters: Map<string, { gutter: Extension; items: DecorationMapByLine }> = new Map()
    private reset: number | null = null

    constructor(private readonly view: EditorView) {}

    public update(update: ViewUpdate): void {
        const currentDecorations = update.state.facet(showTextDocumentDecorations)
        const currentEnabledColumnView = update.state.facet(enableExtensionsDecorationsColumnView)
        const isDarkTheme = update.state.facet(EditorView.darkTheme)

        if (
            update.startState.facet(showTextDocumentDecorations) !== currentDecorations ||
            update.startState.facet(enableExtensionsDecorationsColumnView) !== currentEnabledColumnView
        ) {
            this.groupedDecorations = groupDecorations(currentDecorations, currentEnabledColumnView)
            this.updateInlineDecorations(this.groupedDecorations.inline, !isDarkTheme)

            if (this.updateGutters()) {
                // We cannot synchronously dispatch another transaction during
                // an update, so we schedule it but also cancel pending
                // transactions should this be called multiple times in a row
                if (this.reset !== null) {
                    window.clearTimeout(this.reset)
                }
                this.reset = window.setTimeout(() => {
                    this.view.dispatch({
                        effects: decorationGutters.reconfigure(
                            Array.from(this.gutters.values(), ({ gutter }) => gutter)
                        ),
                    })
                }, 50)
            }
        } else if (update.viewportChanged || isDarkTheme !== update.startState.facet(EditorView.darkTheme)) {
            this.updateInlineDecorations(this.groupedDecorations.inline, !isDarkTheme)
            // Updating column decorators is handled but the gutter extension itself
        }
    }

    private updateInlineDecorations(decorations: GroupedDecorations['inline'], isLightTheme: boolean): void {
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

    /**
     * Create or remove gutters.
     */
    private updateGutters(): boolean {
        let change = false
        const seen: Set<string> = new Set()

        for (const [{ extensionID }, items] of this.groupedDecorations.column) {
            if (!extensionID) {
                continue
            }

            seen.add(extensionID)
            if (!this.gutters.has(extensionID)) {
                const className = extensionID.replace(/\//g, '-')

                this.gutters.set(extensionID, {
                    gutter: gutter({
                        class: `${columnDecoratorStyles.decoration} ${className}`,
                        lineMarker: (view, lineBlock) => {
                            const items = this.gutters.get(extensionID)?.items
                            if (!items) {
                                // This shouldn't be possible but just in case
                                return null
                            }
                            const lineItems = items.get(view.state.doc.lineAt(lineBlock.from).number)
                            if (!lineItems || lineItems.length === 0) {
                                return null
                            }
                            return new ColumnDecoratorMarker(lineItems, !view.state.facet(EditorView.darkTheme))
                        },
                        // Without a spacer the whole gutter flickers when the
                        // decorations for the visible lines are re-rendered
                        // TODO: update spacer when decorations change
                        initialSpacer: () => {
                            const decorations = longestColumnDecorations(this.gutters.get(extensionID)?.items)
                            return new ColumnDecoratorMarker(decorations, /* value doesn't matter for spacer */ true)
                        },
                        // Markers need to be updated when theme changes
                        lineMarkerChange: update =>
                            update.startState.facet(EditorView.darkTheme) !== update.state.facet(EditorView.darkTheme),
                    }),
                    items,
                })
                change = true
            } else {
                this.gutters.get(extensionID)!.items = items
            }
        }

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
 * Used to find the decoration(s) with the longest text, so that they can be
 * used as gutter spacer.
 */
function longestColumnDecorations(mappedDecorations: DecorationMapByLine | undefined): TextDocumentDecoration[] {
    if (!mappedDecorations) {
        return []
    }

    let longest = 0
    let result: TextDocumentDecoration[] = []

    for (const decorations of mappedDecorations.values()) {
        const size = decorations.reduce((size, decoration) => size + (decoration.after?.contentText?.length ?? 0), 0)
        if (size > longest) {
            longest = size
            result = decorations
        }
    }

    return result
}

/**
 * Widget class for rendering column Sourcegrpah text document decorations inside
 * CodeMirror.
 */
class ColumnDecoratorMarker extends GutterMarker {
    private container: HTMLElement | null = null
    private reactRoot: Root | null = null

    constructor(public readonly items: TextDocumentDecoration[], public readonly isLightTheme: boolean) {
        super()
    }

    /* eslint-disable-next-line id-length*/
    public eq(other: ColumnDecoratorMarker): boolean {
        // TODO: Find out why gutter markers still flicker when gutter is
        // redrawn
        return isEqual(this.items, other.items) && this.isLightTheme === other.isLightTheme
    }

    public toDOM(): HTMLElement {
        if (!this.container) {
            this.container = document.createElement('span')
            this.reactRoot = createRoot(this.container)
            this.reactRoot.render(
                <ColumnDecoratorContents lineDecorations={this.items} isLightTheme={this.isLightTheme} />
            )
        }
        return this.container
    }

    public destroy(): void {
        this.container?.remove()
        // setTimeout seems necessary to prevent React from complaining that the
        // root is synchronously unmounted while rendering is in progress
        setTimeout(() => this.reactRoot?.unmount(), 0)
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

// Needed to make sure column and inline decorations are rendered correctly
const columnTheme = EditorView.theme({
    [`.${columnDecoratorStyles.decoration} a`]: {
        width: '100%',
    },
    [`.${lineDecoratorStyles.contents}::before`]: {
        content: 'attr(data-contents)',
    },
})

/**
 * Facet to allow extensions to provide Sourcegraph text document decorations.
 */
export const showTextDocumentDecorations = Facet.define<TextDocumentDecorationSpec[], TextDocumentDecorationSpec[]>({
    combine: decorations => decorations.flat(),
    compareInput: (a, b) => a === b || (a.length === 0 && b.length === 0),
    enables: [
        Prec.lowest(enableExtensionsDecorationsColumnView.of(false)),
        ViewPlugin.fromClass(TextDocumentDecorationManager, {
            decorations: manager => manager.inlineDecorations,
        }),
        decorationGutters.of([]),
        columnTheme,
    ],
})
