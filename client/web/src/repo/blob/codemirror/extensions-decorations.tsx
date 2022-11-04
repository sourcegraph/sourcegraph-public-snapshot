/**
 * This file contains CodeMirror extensions for rendering Sourcegraph extensions specific
 * text document decorations to CodeMirror decorations. Text document
 * decorations are provided via the {@link showTextDocumentDecorations} facet.
 */
import { Facet, RangeSetBuilder } from '@codemirror/state'
import {
    Decoration,
    DecorationSet,
    EditorView,
    PluginValue,
    ViewPlugin,
    ViewUpdate,
    WidgetType,
} from '@codemirror/view'
import { isEqual } from 'lodash'
import { createRoot, Root } from 'react-dom/client'

import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { DecorationMapByLine, groupDecorationsByLine } from '@sourcegraph/shared/src/api/extension/api/decorations'

import { LineDecoratorContents } from '../LineDecorator'

import lineDecoratorStyles from '../LineDecorator.module.scss'

/**
 * This class manages inline text document decorations.
 * Inline decorations are rendered as widget decorations.
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

    private updateDecorations(decorations: TextDocumentDecoration[], isLightTheme: boolean): void {
        this.decorations = groupDecorationsByLine(decorations)
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
 * Widget class for rendering inline Sourcegrpah text document decorations inside CodeMirror.
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
 * Facet to allow extensions to provide Sourcegraph text document decorations.
 */
export const showTextDocumentDecorations = Facet.define<TextDocumentDecoration[], TextDocumentDecoration[]>({
    combine: decorations => decorations.flat(),
    compareInput: (a, b) => a === b || (a.length === 0 && b.length === 0),
    enables: [
        ViewPlugin.fromClass(TextDocumentDecorationManager, {
            decorations: manager => manager.inlineDecorations,
        }),
        EditorView.theme({
            [`.${lineDecoratorStyles.contents}::before`]: {
                content: 'attr(data-contents)',
            },
        }),
    ],
})
