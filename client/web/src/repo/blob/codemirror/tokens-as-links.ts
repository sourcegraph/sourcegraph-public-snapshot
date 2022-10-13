import { Extension, Facet, RangeSetBuilder, StateEffectType } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { History } from 'history'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { concatMap, debounceTime, map } from 'rxjs/operators'
import { DeepNonNullable } from 'utility-types'

import { logger, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { toPrettyBlobURL, UIRange } from '@sourcegraph/shared/src/util/url'

import { DefinitionResponse, fetchDefinitionsFromRanges } from '../backend'
import { BlobInfo } from '../Blob'

import { SelectedLineRange, selectedLines } from './linenumbers'

interface TokenLink {
    range: UIRange
    url: string
}

class TokensAsLinks implements PluginValue {
    private referencesLinks: TokenLink[]
    private subscription: Subscription | undefined
    private visiblePositionRange: Subject<{ from: number; to: number }> | undefined

    constructor(private view: EditorView, private setTokenLinks: StateEffectType<TokenLink[]>) {
        this.referencesLinks = this.buildReferencesLinksFromStencilRange(view)
        requestAnimationFrame(() => {
            this.view.dispatch({
                effects: this.setTokenLinks.of(this.referencesLinks),
            })
        })

        if (view.state.facet(tokensAsLinks).preloadGoToDefinition) {
            this.visiblePositionRange = new Subject()
            this.subscription = this.visiblePositionRange
                .pipe(
                    debounceTime(200),
                    concatMap(({ from, to }) => this.getDefinitionsLinksWithinRange(view, from, to))
                )
                .subscribe(definitionLinks => {
                    const mergedLinks = this.mergeLinks(this.referencesLinks, definitionLinks)
                    requestAnimationFrame(() => {
                        this.view.dispatch({
                            effects: this.setTokenLinks.of(mergedLinks),
                        })
                    })
                })

            // Trigger first update
            this.visiblePositionRange.next({ from: view.viewport.from, to: view.viewport.to })
        }
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
                effects: this.setTokenLinks.of([]),
            })
        })
    }

    /**
     * Merge two sets of TokenLinks, replacing the url of matching ranges.
     * This is intended to allow us to enhance the basic references links with definitions urls when they are available.
     */
    private mergeLinks(baseLinks: TokenLink[], definitionLinks: TokenLink[]): TokenLink[] {
        const mergedLinks: TokenLink[] = []

        for (const base of baseLinks) {
            const definition = definitionLinks.find(
                def =>
                    def.range.start.line === base.range.start.line &&
                    def.range.start.character === base.range.start.character &&
                    def.range.end.line === base.range.end.line &&
                    def.range.end.character === base.range.end.character
            )

            if (definition) {
                mergedLinks.push(definition)
            } else {
                mergedLinks.push(base)
            }
        }

        return mergedLinks
    }

    /**
     * Build a set of TokenLinks from the stencil ranges in the current viewport.
     * TODO: Do for each update?
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

    /**
     * Fetch definitions for all stencil ranges within the viewport.
     */
    private getDefinitionsLinksWithinRange(
        view: EditorView,
        lineFrom: number,
        lineTo: number
    ): Observable<TokenLink[]> {
        const {
            blobInfo: { stencil, repoName, revision, filePath, commitID },
        } = view.state.facet(tokensAsLinks)

        if (!stencil) {
            return of()
        }

        const from = view.state.doc.lineAt(lineFrom).number
        const to = view.state.doc.lineAt(lineTo).number

        // We only want to fetch ranges that are within the current viewport
        const ranges = stencil.filter(({ start, end }) => start.line + 1 >= from && end.line + 1 <= to)

        /**
         * Fetch definition information for all known stencil ranges within the viewport
         * TODO: We should expose an API endpoint that does this rather than batching from the client.
         * GitHub issue: UPDATEME
         */
        return fetchDefinitionsFromRanges({ ranges, repoName, filePath, revision }).pipe(
            map(results =>
                results
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
            )
        )
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
 * Given a set of TokenLinks, returns a decoration set that wraps each specified range in a `<a>` tag.
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
    combine: links => links.flat(),
    enables: facet =>
        EditorView.decorations.compute([facet], state => {
            const links = state.facet(facet)
            let decorations: DecorationSet | null = null

            return view => {
                if (decorations) {
                    return decorations
                }

                return (decorations = tokenLinksToRangeSet(view, links))
            }
        }),
})

function tokenLinks(): Extension {
    const [tokenLinksField, , setTokenLinks] = createUpdateableField<TokenLink[]>([], field =>
        decorateTokensAsLinks.from(field)
    )

    return [
        tokenLinksField,
        focusSelectedLine,
        ViewPlugin.define(view => new TokensAsLinks(view, setTokenLinks)),
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
    preloadGoToDefinition: boolean
}

/**
 * Facet with which we can provide `BlobInfo`, specifically `stencil` ranges.
 *
 * This enables the `tokenLinks` extension which will decorate tokens with links
 * to either their definition or the references panel.
 */
export const tokensAsLinks = Facet.define<TokensAsLinksFacet, TokensAsLinksFacet>({
    combine: source => source[0],
    enables: tokenLinks(),
})
