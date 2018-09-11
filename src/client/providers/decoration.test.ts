import * as assert from 'assert'
import { of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { TextDocumentIdentifier } from '../../client/types/textDocument'
import { TextDocumentDecoration } from '../../protocol/plainTypes'
import { getDecorations, ProvideTextDocumentDecorationSignature } from './decoration'
import { FIXTURE as COMMON_FIXTURE } from './registry.test'

const FIXTURE = {
    ...COMMON_FIXTURE,
    TextDocumentIdentifier: { uri: 'file:///f' } as TextDocumentIdentifier,
}

const FIXTURE_RESULT: TextDocumentDecoration[] | null = [
    {
        range: { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } },
        backgroundColor: 'red',
    },
]

const scheduler = () => new TestScheduler((a, b) => assert.deepStrictEqual(a, b))

describe('getDecorations', () => {
    describe('0 providers', () => {
        it('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-|', { a: [] }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))
    })

    describe('1 provider', () => {
        it('returns null result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-|', { a: [() => of(null)] }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        it('returns result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULT)],
                        }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_RESULT,
                })
            ))
    })

    describe('2 providers', () => {
        it('returns null result if both providers return null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-|', {
                            a: [() => of(null), () => of(null)],
                        }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        it('omits null result from 1 provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULT), () => of(null)],
                        }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_RESULT,
                })
            ))

        it('merges results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULT), () => of(FIXTURE_RESULT)],
                        }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-|', {
                    a: [...FIXTURE_RESULT!, ...FIXTURE_RESULT!],
                })
            ))
    })

    describe('multiple emissions', () => {
        it('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-b-|', {
                            a: [() => of(FIXTURE_RESULT)],
                            b: [() => of(null)],
                        }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-b-|', {
                    a: FIXTURE_RESULT,
                    b: null,
                })
            ))
    })
})
