import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView } from '@codemirror/view'
import { History } from 'history'

import { logger } from '@sourcegraph/common'
import { UIRange } from '@sourcegraph/shared/src/util/url'

interface TokenLink {
    range: UIRange
    url: string
}

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
                    class: 'sourcegraph-token-link',
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

interface TokensAsLinksFacet {
    history: History
    links: TokenLink[]
}

/**
 * Facet with which we can provide TokenLinks and have them rendered as interactive links.
 */
export const tokensAsLinks = Facet.define<TokensAsLinksFacet, TokensAsLinksFacet>({
    combine: source => source[0],
    enables: facet => [
        EditorView.decorations.compute([facet], state => {
            const { links } = state.facet(facet)
            let decorations: DecorationSet | null = null

            return view => {
                if (decorations) {
                    return decorations
                }

                return (decorations = tokenLinksToRangeSet(view, links))
            }
        }),
        EditorView.domEventHandlers({
            click(event: MouseEvent, view: EditorView) {
                const target = event.target as HTMLElement

                // Check to see if the clicked target is a token link.
                // If it is, push the link to the history stack.
                if (target.matches('[data-token-link]')) {
                    event.preventDefault()
                    const { history } = view.state.facet(facet)
                    history.push(target.getAttribute('href')!)
                }
            },
        }),
    ],
})
