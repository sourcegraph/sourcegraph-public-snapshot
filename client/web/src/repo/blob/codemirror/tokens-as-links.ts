import { Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { History } from 'history'

import { logger } from '@sourcegraph/common'
import { UIRange } from '@sourcegraph/shared/src/util/url'

interface TokenLink {
    range: UIRange
    url: string
}

class DefinitionManager implements PluginValue {
    private subscription: Subscription
    private visiblePositionRange: Subject<{ from: number; to: number }>

    constructor(private view: EditorView, private blobInfo: BlobInfo) {
        this.visiblePositionRange = new Subject()
        this.subscription = this.visiblePositionRange
            .pipe(
                debounceTime(200),
                concatMap(({ from, to }) => this.getDefinitionsLinksWithinRange(view, from, to))
            )
            .subscribe(definitionLinks => {
                requestAnimationFrame(() => {
                    this.view.dispatch({
                        effects: setTokenLinks.of(definitionLinks),
                    })
                })
            })

        // Trigger first update
        this.visiblePositionRange.next({ from: view.viewport.from, to: view.viewport.to })
    }

    public update(update: ViewUpdate): void {
        if (update.viewportChanged) {
            this.visiblePositionRange?.next({ from: update.view.viewport.from, to: update.view.viewport.to })
        }
    }

    public destroy(): void {
        this.subscription?.unsubscribe()
        requestAnimationFrame(() => {
            this.view.dispatch({
                effects: setTokenLinks.of([]),
            })
        })
    }

    /**
     * Fetch definitions for all stencil ranges within the viewport.
     */
    private getDefinitionsLinksWithinRange(
        view: EditorView,
        lineFrom: number,
        lineTo: number
    ): Observable<TokenLink[]> {
        const { repoName, filePath, revision, commitID } = this.blobInfo

        const from = view.state.doc.lineAt(lineFrom).number
        const to = view.state.doc.lineAt(lineTo).number

        // We only want to fetch definitions for ranges that:
        // 1. We know are valid (i.e. already set in the state)
        // 2. Are within the current viewport
        const ranges = view.state
            .field(tokenLinks)
            .filter(({ range }) => range.start.line + 1 >= from && range.end.line + 1 <= to)
            .map(({ range }) => range)

        /**
         * Fetch definitions from known ranges, and convert to `TokenLink`s
         */
        return fetchDefinitionsFromRanges({ ranges, repoName, filePath, revision }).pipe(
            map(results => {
                if (results === null) {
                    return []
                }

                return results
                    .filter((result): result is DeepNonNullable<DefinitionResponse> => Boolean(result.definition))
                    .map(({ range, definition }) => {
                        // Preserve the users current revision (e.g., a Git branch name instead of a Git commit SHA)
                        // This avoids navigating the user from (e.g.) a URL with a nice Git
                        // branch name to a URL with a full Git commit SHA.
                        const targetRevision =
                            definition.resource.repository.name === repoName &&
                            definition.resource.commit.oid === commitID
                                ? revision
                                : definition.resource.commit.oid

                        return {
                            range,
                            url: toPrettyBlobURL({
                                repoName: definition.resource.repository.name,
                                filePath: definition.resource.path,
                                revision: targetRevision,
                                position: definition.range
                                    ? {
                                          line: definition.range.start.line + 1,
                                          character: definition.range.start.character,
                                      }
                                    : undefined,
                            }),
                        }
                    })
            })
        )
    }
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
        })
    ]
})

interface TokensAsLinksConfiguration {
    history: History
    blobInfo: BlobInfo
    preloadGoToDefinition: boolean
}

export const tokensAsLinks = ({ history, blobInfo, preloadGoToDefinition }: TokensAsLinksConfiguration): Extension => {
    const referencesLinks =
        blobInfo.stencil?.map(range => ({
            range,
            url: `?${toPositionOrRangeQueryParameter({
                position: { line: range.start.line + 1, character: range.start.character + 1 },
            })}#tab=references`,
        })) ?? []

    return [
        tokenLinks.init(() => referencesLinks),
        preloadGoToDefinition ? ViewPlugin.define(view => new DefinitionManager(view, blobInfo)) : [],
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
