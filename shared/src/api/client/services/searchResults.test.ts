import { SearchResult } from '@sourcegraph/extension-api-types'
import { of, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { provideSearchResult, ProvideSearchResultSignature } from './searchResults'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const FIXTURE_SEARCH_RESULT: SearchResult = {
    label: { text: 'Label', html: '' },
    detail: { text: 'Detail', html: '' },
    icon: '',
    url: 'http://example.com',
    matches: [
        {
            url: 'http://example.com',
            body: { text: 'Result body', html: '' },
            highlights: [{ start: { line: 0, character: 2 }, end: { line: 0, character: 4 } }],
        },
    ],
}

const FIXTURE_SEARCH_RESULTS: SearchResult | SearchResult[] | null = [FIXTURE_SEARCH_RESULT, FIXTURE_SEARCH_RESULT]

describe('provideSearchResult', () => {
    describe('0 providers', () => {
        test('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    provideSearchResult(cold<ProvideSearchResultSignature[]>('-a-|', { a: [] }), 'foo')
                ).toBe('-a-|', {
                    a: null,
                })
            ))
    })

    describe('1 provider', () => {
        test('returns null result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    provideSearchResult(cold<ProvideSearchResultSignature[]>('-a-|', { a: [() => of(null)] }), 'foo')
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        test('returns result array from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    provideSearchResult(
                        cold<ProvideSearchResultSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_SEARCH_RESULTS)],
                        }),
                        'fpp'
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_SEARCH_RESULTS,
                })
            ))
    })

    test('errors do not propagate', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                provideSearchResult(
                    cold<ProvideSearchResultSignature[]>('-a-|', {
                        a: [() => of([FIXTURE_SEARCH_RESULT]), () => throwError('x')],
                    }),
                    'foo',
                    false
                )
            ).toBe('-a-|', {
                a: [FIXTURE_SEARCH_RESULT],
            })
        ))

    describe('2 providers', () => {
        test('returns null result if both providers return null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    provideSearchResult(
                        cold<ProvideSearchResultSignature[]>('-a-|', {
                            a: [() => of(null), () => of(null)],
                        }),
                        'foo'
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        test('omits null result from 1 provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    provideSearchResult(
                        cold<ProvideSearchResultSignature[]>('-a-|', {
                            a: [() => of([FIXTURE_SEARCH_RESULT]), () => of(null)],
                        }),
                        'foo'
                    )
                ).toBe('-a-|', {
                    a: [FIXTURE_SEARCH_RESULT],
                })
            ))

        test('merges results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    provideSearchResult(
                        cold<ProvideSearchResultSignature[]>('-a-|', {
                            a: [() => of([FIXTURE_SEARCH_RESULT]), () => of([FIXTURE_SEARCH_RESULT])],
                        }),
                        'foo'
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_SEARCH_RESULTS,
                })
            ))
    })

    describe('multiple emissions', () => {
        test('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    provideSearchResult(
                        cold<ProvideSearchResultSignature[]>('-a-b-|', {
                            a: [() => of([FIXTURE_SEARCH_RESULT])],
                            b: [() => of(null)],
                        }),
                        'foo'
                    )
                ).toBe('-a-b-|', {
                    a: [FIXTURE_SEARCH_RESULT],
                    b: null,
                })
            ))
    })
})
