import * as assert from 'assert'
import { of, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { Definition, Position, Range } from 'vscode-languageserver-types'
import { getDefinition, ProvideTextDocumentDefinitionSignature } from './definition'
import { FIXTURE } from './textDocument.test'

const scheduler = () => new TestScheduler((a, b) => assert.deepStrictEqual(a, b))

const FIXTURE_RESULT: Definition | null = [
    { uri: 'file:///f', range: Range.create(Position.create(1, 2), Position.create(3, 4)) },
]

describe('getDefinition', () => {
    describe('0 providers', () => {
        it('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-|', { a: [] }),
                        FIXTURE.TextDocumentPositionParams
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
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-|', { a: [() => of(null)] }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        it('returns result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULT)],
                        }),
                        FIXTURE.TextDocumentPositionParams
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
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-|', {
                            a: [() => of(null), () => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        it('omits null result from 1 provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULT), () => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_RESULT,
                })
            ))

        it('skips errors from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULT), () => throwError('error')],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_RESULT,
                })
            ))

        it('merges results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-|', {
                            a: [
                                () =>
                                    of({
                                        uri: 'file:///f1',
                                        range: { start: Position.create(1, 2), end: Position.create(3, 4) },
                                    }),
                                () =>
                                    of({
                                        uri: 'file:///f2',
                                        range: { start: Position.create(5, 6), end: Position.create(7, 8) },
                                    }),
                            ],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-|', {
                    a: [
                        {
                            uri: 'file:///f1',
                            range: { start: Position.create(1, 2), end: Position.create(3, 4) },
                        },
                        {
                            uri: 'file:///f2',
                            range: { start: Position.create(5, 6), end: Position.create(7, 8) },
                        },
                    ],
                })
            ))
    })

    describe('multiple emissions', () => {
        it('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDefinition(
                        cold<ProvideTextDocumentDefinitionSignature[]>('-a-b-|', {
                            a: [() => of(FIXTURE_RESULT)],
                            b: [() => of(null)],
                        }),
                        FIXTURE.TextDocumentPositionParams
                    )
                ).toBe('-a-b-|', {
                    a: FIXTURE_RESULT,
                    b: null,
                })
            ))
    })
})
