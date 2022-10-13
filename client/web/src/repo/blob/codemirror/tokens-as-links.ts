import { Extension, Facet, RangeSetBuilder, StateEffectType } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { History } from 'history'

import { logger, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { UIRange } from '@sourcegraph/shared/src/util/url'

import { BlobInfo } from '../Blob'

import { SelectedLineRange, selectedLines } from './linenumbers'

interface TokenLink {
    range: UIRange
    url: string
}

class TokensAsLinks implements PluginValue {
    constructor(private view: EditorView, private setTokenLinks: StateEffectType<TokenLink[]>) {
        const referencesLinks = this.buildReferencesLinksFromStencilRange(view)
        requestAnimationFrame(() => {
            this.view.dispatch({
                effects: this.setTokenLinks.of(referencesLinks),
            })
        })
    }

    public destroy(): void {
        requestAnimationFrame(() => {
            this.view.dispatch({
                effects: this.setTokenLinks.of([]),
            })
        })
    }

    /**
     * Build a set of TokenLinks from a set of stencil ranges
     */
    private buildReferencesLinksFromStencilRange(view: EditorView): TokenLink[] {
        const {
            blobInfo: { stencil },
        } = view.state.facet(tokensAsLinks)

        if (!stencil) {
            return []
        }

        return stencil.map(range => ({
            range,
            url: `?${toPositionOrRangeQueryParameter({
                position: { line: range.start.line + 1, character: range.start.character + 1 },
            })}#tab=references`,
        }))
    }
}

/**
 * View plugin responsible for focusing a selected line when necessary.
 */
const focusSelectedLine = ViewPlugin.fromClass(
    class implements PluginValue {
        private lastSelectedLines: SelectedLineRange | null = null
        constructor(private readonly view: EditorView) {}

        public update(update: ViewUpdate): void {
            const currentSelectedLines = update.state.field(selectedLines)
            if (this.lastSelectedLines !== currentSelectedLines) {
                this.lastSelectedLines = currentSelectedLines
                this.focusLine(currentSelectedLines)
            }
        }

        public focusLine(selection: SelectedLineRange): void {
            if (selection) {
                const line = this.view.state.doc.line(selection.line)

                if (line) {
                    window.requestAnimationFrame(() => {
                        const element = this.view.domAtPos(line.from).node as HTMLElement
                        element.focus()
                    })
                }
            }
        }
    }
)

/**
 * Given a set of ranges, returns a decoration set that adds a link to each range.
 */
function tokenLinksToRangeSet(view: EditorView, links: TokenLink[]): DecorationSet {
    const builder = new RangeSetBuilder<Decoration>()

    try {
        for (const { range, url } of links) {
            const from = view.state.doc.line(range.start.line + 1).from + range.start.character
            const to = view.state.doc.line(range.end.line + 1).from + range.end.character
            const decoration = Decoration.mark({
                attributes: {
                    class: 'sourcegraph-document-focus',
                    href: url,
                    'data-token-link': '',
                },
                tagName: 'a',
            })
            builder.add(from, to, decoration)
        }
    } catch (error) {
        logger.error('Failed to compute decorations', error)
    }

    return builder.finish()
}

const decorateTokensAsLinks = Facet.define<TokenLink[], TokenLink[]>({
    combine: ranges => ranges.flat(),
    enables: facet =>
        EditorView.decorations.compute([facet], state => {
            const ranges = state.facet(facet)
            let decorations: DecorationSet | null = null

            return view => {
                if (decorations) {
                    return decorations
                }

                return (decorations = tokenLinksToRangeSet(view, ranges))
            }
        }),
})

function tokenLinks(): Extension {
    const [tokenLinksField, , setTokenLinks] = createUpdateableField<TokenLink[]>([], field =>
        decorateTokensAsLinks.from(field)
    )

    return [
        tokenLinksField,
        ViewPlugin.define(view => new TokensAsLinks(view, setTokenLinks)),
        focusSelectedLine,
        EditorView.domEventHandlers({
            click(event: MouseEvent, view: EditorView) {
                const target = event.target as HTMLElement

                // Check to see if the clicked target is a token link.
                // If it is, push the link to the history stack.
                if (target.matches('[data-token-link]')) {
                    event.preventDefault()
                    const { history } = view.state.facet(tokensAsLinks)
                    history.push(target.getAttribute('href')!)
                }
            },
        }),
    ]
}

interface TokensAsLinksFacet {
    blobInfo: BlobInfo
    history: History
}

/**
 * Facet with which we can provide `BlobInfo`, specifically `stencil` ranges.
 *
 * This enables the `tokenLinks` extension which will decorate tokens with links to the references panel
 */
export const tokensAsLinks = Facet.define<TokensAsLinksFacet, TokensAsLinksFacet>({
    combine: source => source[0],
    enables: tokenLinks(),
})
