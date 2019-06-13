import { of, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { TextSearchResult } from 'sourcegraph'
import { getResults, ProvideTextSearchResultsParams, ProvideTextSearchResultsSignature } from './searchProviders'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const FIXTURE_PARAMS: ProvideTextSearchResultsParams = {
    query: { pattern: 'p', type: 'regexp' },
    options: { files: { includes: ['p'], type: 'regexp' } },
}

const FIXTURE_RESULTS: TextSearchResult[] = [{ uri: 'file:///f0' }, { uri: 'file:///f1' }]

describe('getResults', () => {
    describe('0 providers', () => {
        test('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResults(cold<ProvideTextSearchResultsSignature[]>('-a-|', { a: [] }), FIXTURE_PARAMS)
                ).toBe('-a-|', {
                    a: cold<TextSearchResult[]>('(a|)', { a: [] }),
                })
            ))
    })

    describe('1 provider', () => {
        test('returns empty result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResults(cold<ProvideTextSearchResultsSignature[]>('-a-|', { a: [() => of([])] }), FIXTURE_PARAMS)
                ).toBe('-a-|', {
                    a: cold<TextSearchResult[]>('(a|)', { a: [] }),
                })
            ))

        test('returns result array from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResults(
                        cold<ProvideTextSearchResultsSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULTS)],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: cold<TextSearchResult[]>('(a|)', { a: FIXTURE_RESULTS }),
                })
            ))
    })

    test('errors do not propagate', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getResults(
                    cold<ProvideTextSearchResultsSignature[]>('-a-|', {
                        a: [() => of(FIXTURE_RESULTS), () => throwError(new Error('x'))],
                    }),
                    FIXTURE_PARAMS,
                    false
                )
            ).toBe('-a-|', {
                a: cold<TextSearchResult[]>('(a|)', { a: FIXTURE_RESULTS }),
            })
        ))

    describe('2 providers', () => {
        test('returns 1 empty result if both providers return empty', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResults(
                        cold<ProvideTextSearchResultsSignature[]>('-a-|', {
                            a: [() => of([]), () => of([])],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: cold<TextSearchResult[]>('(a|)', { a: [] }),
                })
            ))

        test('omits empty result from 1 provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResults(
                        cold<ProvideTextSearchResultsSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULTS), () => of([])],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: cold<TextSearchResult[]>('(a|)', { a: FIXTURE_RESULTS }),
                })
            ))

        test('emits results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResults(
                        cold<ProvideTextSearchResultsSignature[]>('-a-|', {
                            a: [() => of([{ uri: 'file:///f1' }]), () => of([{ uri: 'file:///f2' }])],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: cold<TextSearchResult[]>('(ab|)', {
                        a: [{ uri: 'file:///f1' }],
                        b: [{ uri: 'file:///f2' }],
                    }),
                })
            ))
    })

    describe('multiple emissions', () => {
        test('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getResults(
                        cold<ProvideTextSearchResultsSignature[]>('-a-b-|', {
                            a: [() => of(FIXTURE_RESULTS)],
                            b: [() => of([])],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-b-|', {
                    a: cold<TextSearchResult[]>('(a|)', { a: FIXTURE_RESULTS }),
                    b: cold<TextSearchResult[]>('(a|)', { a: [] }),
                })
            ))
    })
})
