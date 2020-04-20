import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { TestScheduler } from 'rxjs/testing'
import { Hover } from 'sourcegraph'
import { HoverMerged } from '../types/hover'
import { getHover, ProvideTextDocumentHoverSignature } from './hover'
import { FIXTURE } from './registry.test'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

const FIXTURE_RESULT: Hover | null = { contents: { value: 'c', kind: MarkupKind.PlainText } }
const FIXTURE_RESULT_MERGED: HoverMerged | null = { contents: [{ value: 'c', kind: MarkupKind.PlainText }] }

describe('getHover', () => {
    describe('0 providers', () => {
        test('returns null', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a', { a: [] }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a', {
                    a: { isLoading: false, result: null },
                })
            )
        })
    })

    describe('1 provider', () => {
        it('returns null result from provider', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a', { a: [() => cold('--a', { a: null })] }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-l-r', {
                    l: { isLoading: true, result: null },
                    r: { isLoading: false, result: null },
                })
            )
        })

        test('returns result from provider', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a', {
                            a: [() => cold('-a', { a: FIXTURE_RESULT })],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-lr', {
                    l: { isLoading: true, result: null },
                    r: { isLoading: false, result: FIXTURE_RESULT_MERGED },
                })
            )
        })
    })

    describe('2 providers', () => {
        it('returns null result if both providers return null', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a', {
                            a: [() => cold('-a', { a: null }), () => cold('-a', { a: null })],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-lr', {
                    l: { isLoading: true, result: null },
                    r: { isLoading: false, result: null },
                })
            )
        })

        it('omits null result from 1 provider', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a', {
                            a: [() => cold('-a', { a: FIXTURE_RESULT }), () => cold('-a', { a: null })],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-lr', {
                    l: { isLoading: true, result: null },
                    r: { isLoading: false, result: FIXTURE_RESULT_MERGED },
                })
            )
        })

        it('omits error result from 1 provider', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a', {
                            a: [() => cold('-a', { a: FIXTURE_RESULT }), () => cold('-#', {}, new Error('err'))],
                        }),
                        FIXTURE.TextDocumentPositionParams,
                        false
                    )
                ).toBe('-lr', {
                    l: { isLoading: true, result: null },
                    r: { isLoading: false, result: FIXTURE_RESULT_MERGED },
                })
            )
        })

        it('merges results from providers', () => {
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a-|', {
                            a: [
                                () =>
                                    cold('-a', {
                                        a: {
                                            contents: { value: 'c1' },
                                            range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                                        },
                                    }),
                                () =>
                                    cold('-a', {
                                        a: {
                                            contents: { value: 'c2' },
                                            range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                                        },
                                    }),
                            ],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-lr', {
                    l: { isLoading: true, result: null },
                    r: {
                        isLoading: false,
                        result: {
                            contents: [
                                { value: 'c1', kind: MarkupKind.PlainText },
                                { value: 'c2', kind: MarkupKind.PlainText },
                            ],
                            range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
                        },
                    },
                })
            )
        })
    })

    describe('multiple emissions', () => {
        it('returns stream of results', () => {
            scheduler().run(({ cold, expectObservable }) => {
                expectObservable(
                    getHover(
                        cold<ProvideTextDocumentHoverSignature[]>('-a-b', {
                            a: [() => cold('-a', { a: FIXTURE_RESULT })],
                            b: [() => cold('-a', { a: null })],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-abcd', {
                    a: { isLoading: true, result: null },
                    b: { isLoading: false, result: FIXTURE_RESULT_MERGED },
                    c: { isLoading: true, result: null },
                    d: { isLoading: false, result: null },
                })
            })
        })
    })
})
