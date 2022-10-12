import { Extension, Facet, RangeSetBuilder, StateEffectType } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, PluginValue, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { History } from 'history'
import { Observable, of, Subject, Subscription, zip } from 'rxjs'
import { concatMap, map } from 'rxjs/operators'

import { logger, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { toPrettyBlobURL, UIRange } from '@sourcegraph/shared/src/util/url'

import { DefinitionFields } from '../../../graphql-operations'
import { fetchDefinition } from '../backend'
import { BlobInfo } from '../Blob'

import { SelectedLineRange, selectedLines } from './linenumbers'

interface TokenLink {
    from: number
    to: number
    url: string
}

interface FilteredDefinitionResponse {
    range: UIRange
    definition: DefinitionFields
}

class TokensAsLinks implements PluginValue {
    private stencilRanges: TokenLink[]
    private subscription: Subscription | undefined
    private visiblePositionRange: Subject<{ from: number; to: number }> | undefined

    constructor(private view: EditorView, private setTokenLinkRanges: StateEffectType<TokenLink[]>) {
        this.stencilRanges = this.buildBasicLinksFromStencilRange(view)
        requestAnimationFrame(() => {
            this.view.dispatch({
                effects: this.setTokenLinkRanges.of(this.stencilRanges),
            })
        })

        if (view.state.facet(tokensAsLinks).preloadGoToDefinition) {
            this.visiblePositionRange = new Subject()
            this.subscription = this.visiblePositionRange
                .pipe(concatMap(({ from, to }) => this.getDefinitionsLinksWithinRange(view, from, to)))
                .subscribe(definitions => {
                    const mergedRanges = this.mergeRanges(this.stencilRanges, definitions)
                    requestAnimationFrame(() => {
                        this.view.dispatch({
                            effects: this.setTokenLinkRanges.of(mergedRanges),
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
                effects: this.setTokenLinkRanges.of([]),
            })
        })
    }

    /**
     * Merge two sets of TokenLink ranges, replacing the url of any overlapping ranges.
     * This is intended to allow us to enhance the stencil ranges with definitions urls when they are available.
     */
    private mergeRanges(baseRanges: TokenLink[], additionalRanges: TokenLink[]): TokenLink[] {
        const mergedRanges: TokenLink[] = []

        for (const range of baseRanges) {
            const definition = additionalRanges.find(
                additionalRange => additionalRange.from === range.from && additionalRange.to === range.to
            )

            if (definition) {
                mergedRanges.push(definition)
            } else {
                mergedRanges.push(range)
            }
        }

        return mergedRanges
    }

    /**
     * Build a set of TokenLink ranges from the stencil ranges in the current viewport.
     */
    private buildBasicLinksFromStencilRange(view: EditorView): TokenLink[] {
        const {
            blobInfo: { stencil },
        } = view.state.facet(tokensAsLinks)

        if (!stencil) {
            return []
        }

        return stencil.map(({ start, end }) => ({
            from: this.view.state.doc.line(start.line + 1).from + start.character,
            to: this.view.state.doc.line(end.line + 1).from + end.character,
            url: `?${toPositionOrRangeQueryParameter({
                position: { line: start.line + 1, character: start.character + 1 },
            })}#tab=references`,
        }))
    }

    /**
     * Given a set of stencil ranges within a viewport,
     * fetch the definition for each range and return a set of TokenLink ranges.
     */
    private getDefinitionsLinksWithinRange(
        view: EditorView,
        lineFrom: number,
        lineTo: number
    ): Observable<TokenLink[]> {
        const {
            blobInfo: { stencil, repoName, revision, filePath },
        } = view.state.facet(tokensAsLinks)

        if (!stencil) {
            return of()
        }

        const from = view.state.doc.lineAt(lineFrom).number
        const to = view.state.doc.lineAt(lineTo).number

        /**
         * Fetch definition information for all known stencil ranges within the viewport
         * TODO: We should expose an API endpoint that does this rather than firing a load of individual requests
         * GitHub issue: UPDATEME
         */
        return zip(
            ...stencil
                .filter(({ start, end }) => start.line + 1 >= from && end.line + 1 <= to)
                .map(({ start, end }) =>
                    fetchDefinition({
                        repoName,
                        filePath,
                        revision,
                        position: start,
                    }).pipe(
                        map(definition => ({
                            range: { start, end },
                            definition,
                        }))
                    )
                )
        ).pipe(
            map(results =>
                results
                    .filter((result): result is FilteredDefinitionResponse => Boolean(result.definition))
                    .map(({ range, definition }) => {
                        const { start, end } = range
                        const url = toPrettyBlobURL({
                            repoName: definition.resource.repository.name,
                            filePath: definition.resource.path,
                            revision,
                            commitID: definition.resource.commit.oid,
                            position: definition.range
                                ? {
                                      line: definition.range.start.line + 1,
                                      character: definition.range.start.character,
                                  }
                                : undefined,
                        })

                        return {
                            from: view.state.doc.line(start.line + 1).from + start.character,
                            to: view.state.doc.line(end.line + 1).from + end.character,
                            url,
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
 * Given a set of ranges, returns a decoration set that adds a link to each range.
 */
function tokenLinksToRangeSet(ranges: TokenLink[]): DecorationSet {
    const builder = new RangeSetBuilder<Decoration>()

    try {
        for (const { from, to, url } of ranges) {
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

            return () => {
                if (decorations) {
                    return decorations
                }

                return (decorations = tokenLinksToRangeSet(ranges))
            }
        }),
})

function tokenLinks(): Extension {
    const [tokenRangesField, , setTokenRanges] = createUpdateableField<TokenLink[]>([], field =>
        decorateTokensAsLinks.from(field)
    )

    return [
        tokenRangesField,
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
        focusSelectedLine,
        ViewPlugin.define(view => new TokensAsLinks(view, setTokenRanges)),
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
