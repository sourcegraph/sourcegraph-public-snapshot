import type { Observable } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { describe, expect, it } from 'vitest'

import { fromHoverMerged, type HoverMerged, type TextDocumentIdentifier } from '@sourcegraph/client-api'
import { LOADING } from '@sourcegraph/codeintellify'
import { MarkupKind, Range } from '@sourcegraph/extension-api-classes'

import type { Hover, DocumentHighlight } from '../../../codeintel/legacy-extensions/api'
import { callProviders, mergeProviderResults, providersForDocument, type RegisteredProvider } from '../extensionHostApi'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

type Provider = RegisteredProvider<number | Observable<number>>

const documentURI = 'git://repo#src/f.ts'

const getResultsFromProviders = (providersObservable: Observable<Provider[]>, document: TextDocumentIdentifier) =>
    callProviders(
        providersObservable,
        providers => providersForDocument(document, providers, ({ selector }) => selector),
        ({ provider }) => provider,
        results => results,
        false // < -- logErrors
    )

describe('callProviders()', () => {
    const provide = (number: number | Observable<number>, pattern = '*.ts'): Provider => ({
        provider: number,
        selector: [{ pattern }],
    })

    describe('1 provider', () => {
        it('returns empty non loading result with no providers', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(getResultsFromProviders(cold<Provider[]>('-a', { a: [] }), { uri: documentURI })).toBe(
                    '-a',
                    {
                        a: { isLoading: false, result: [] },
                    }
                )
            )
        })

        it('returns a result from a provider synchronously with raw values', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResultsFromProviders(cold<Provider[]>('-a', { a: [provide(1)] }), { uri: documentURI })
                ).toBe('-(lr)', {
                    l: { isLoading: true, result: [LOADING] },
                    r: { isLoading: false, result: [1] },
                })
            )
        })

        it('returns a result from a provider with observables', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResultsFromProviders(cold<Provider[]>('-a', { a: [provide(cold('--a', { a: 1 }))] }), {
                        uri: documentURI,
                    })
                ).toBe('-l-r', {
                    l: { isLoading: true, result: [LOADING] },
                    r: { isLoading: false, result: [1] },
                })
            )
        })
    })

    describe('2 providers', () => {
        it("returns a synchronous result from both providers, but doesn't wait for the second to yield", () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResultsFromProviders(cold<Provider[]>('-a', { a: [provide(1), provide(2)] }), {
                        uri: documentURI,
                    })
                ).toBe('-(lr)', {
                    l: { isLoading: true, result: [1, LOADING] },
                    r: { isLoading: false, result: [1, 2] },
                })
            )
        })

        it('returns LOADING result first if providers return observables', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResultsFromProviders(
                        cold<Provider[]>('-a', {
                            a: [provide(cold('-a', { a: 1 })), provide(cold('-a', { a: 2 }))],
                        }),
                        { uri: documentURI }
                    )
                ).toBe('-lr', {
                    l: { isLoading: true, result: [LOADING, LOADING] },
                    r: { isLoading: false, result: [1, 2] },
                })
            )
        })

        it('replaces errors from a provider with nulls', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResultsFromProviders(
                        cold<Provider[]>('-a', {
                            a: [provide(cold('-a', { a: 1 })), provide(cold('-#', {}, new Error('boom!')))],
                        }),
                        { uri: documentURI }
                    )
                ).toBe('-lr', {
                    l: { isLoading: true, result: [LOADING, LOADING] },
                    r: { isLoading: false, result: [1, null] },
                })
            )
        })
    })

    describe('filtering', () => {
        it('it can filter out providers', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResultsFromProviders(cold<Provider[]>('-a', { a: [provide(1, '*.ts'), provide(2, '*.js')] }), {
                        uri: documentURI,
                    })
                ).toBe('-(lr)', {
                    l: { isLoading: true, result: [LOADING] },
                    r: { isLoading: false, result: [1] },
                })
            )
        })
    })

    describe('providers change over time', () => {
        it('requeries providers if they changed', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResultsFromProviders(
                        cold<Provider[]>('-a-b', {
                            a: [provide(cold('-a', { a: 1 })), provide(cold('-a', { a: 2 }))],
                            b: [provide(cold('-a', { a: 3 }))],
                        }),
                        { uri: documentURI }
                    )
                ).toBe('-abcd', {
                    a: { isLoading: true, result: [LOADING, LOADING] },
                    b: { isLoading: false, result: [1, 2] },
                    c: { isLoading: true, result: [LOADING] },
                    d: { isLoading: false, result: [3] },
                })
            )
        })
    })
})

describe('mergeProviderResults()', () => {
    describe('document highlight results', () => {
        function mergeDocumentHighlightResults(results: (typeof LOADING | DocumentHighlight[] | null | undefined)[]) {
            return mergeProviderResults(results)
        }

        const range1 = new Range(1, 2, 3, 4)
        const range2 = new Range(2, 3, 4, 5)
        const range3 = new Range(3, 4, 5, 6)

        it('merges non DocumentHighlight values into empty arrays', () => {
            expect(mergeDocumentHighlightResults([LOADING])).toStrictEqual([])
            expect(mergeDocumentHighlightResults([null])).toStrictEqual([])
            expect(mergeDocumentHighlightResults([undefined])).toStrictEqual([])
            // and yes, there can be several
            expect(mergeDocumentHighlightResults([null, LOADING])).toStrictEqual([])
        })

        it('merges a DocumentHighlight into result', () => {
            const highlight1: DocumentHighlight = { range: range1 }
            const highlight2: DocumentHighlight = { range: range2 }
            const highlight3: DocumentHighlight = { range: range3 }
            const merged: DocumentHighlight[] = [highlight1, highlight2, highlight3]
            expect(mergeDocumentHighlightResults([[highlight1], [highlight2, highlight3]])).toEqual(merged)
        })

        it('omits non DocumentHighlight values from document highlight result', () => {
            const highlight: DocumentHighlight = { range: range1 }
            const merged: DocumentHighlight[] = [highlight]
            expect(mergeDocumentHighlightResults([[highlight], null, LOADING, undefined])).toEqual(merged)
        })
    })

    describe('hover results', () => {
        function mergeHoverResults(results: (typeof LOADING | Hover | null | undefined)[]) {
            return fromHoverMerged(mergeProviderResults(results))
        }

        it('merges non Hover values into nulls', () => {
            expect(mergeHoverResults([LOADING])).toBe(null)
            expect(mergeHoverResults([null])).toBe(null)
            expect(mergeHoverResults([undefined])).toBe(null)
            // and yes, there can be several
            expect(mergeHoverResults([null, LOADING])).toBe(null)
        })

        it('merges a Hover into result', () => {
            const hover: Hover = { contents: { value: 'a', kind: MarkupKind.PlainText } }
            const merged: HoverMerged = { contents: [hover.contents], aggregatedBadges: [] }
            expect(mergeHoverResults([hover])).toEqual(merged)
        })

        it('omits non Hover values from hovers result', () => {
            const hover: Hover = { contents: { value: 'a', kind: MarkupKind.PlainText } }
            const merged: HoverMerged = { contents: [hover.contents], aggregatedBadges: [] }
            expect(mergeHoverResults([hover, null, LOADING, undefined])).toEqual(merged)
        })

        it('merges Hovers with ranges', () => {
            const hover1: Hover = {
                contents: { value: 'c1' },
                // TODO this is weird to cast to ranges
                range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } } as unknown as Range,
            }
            const hover2: Hover = {
                contents: { value: 'c2' },
                // TODO this is weird to cast to ranges
                range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } } as unknown as Range,
            }
            const merged: HoverMerged = {
                contents: [
                    { value: 'c1', kind: MarkupKind.PlainText },
                    { value: 'c2', kind: MarkupKind.PlainText },
                ],
                range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                aggregatedBadges: [],
            }

            expect(mergeHoverResults([hover1, hover2])).toEqual(merged)
        })
    })
})
