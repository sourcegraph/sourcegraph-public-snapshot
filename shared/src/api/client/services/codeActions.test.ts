import { Range } from '@sourcegraph/extension-api-classes'
import { of, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import * as sourcegraph from 'sourcegraph'
import { CodeActionsParams, getCodeActions, ProvideCodeActionsSignature } from './codeActions'
import { FIXTURE } from './registry.test'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const FIXTURE_PARAMS: CodeActionsParams = {
    textDocument: FIXTURE.TextDocumentPositionParams.textDocument,
    range: new Range(1, 2, 3, 4),
    context: { diagnostics: [] },
}

const FIXTURE_CODE_ACTION: sourcegraph.CodeAction = {
    title: 'a',
    command: { title: 'c', command: 'c' },
}

const FIXTURE_CODE_ACTIONS: sourcegraph.CodeAction[] = [FIXTURE_CODE_ACTION]

describe('getCodeActions', () => {
    describe('0 providers', () => {
        test('returns null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getCodeActions(cold<ProvideCodeActionsSignature[]>('-a-|', { a: [] }), FIXTURE_PARAMS)
                ).toBe('-a-|', {
                    a: null,
                })
            ))
    })

    describe('1 provider', () => {
        test('returns null result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getCodeActions(cold<ProvideCodeActionsSignature[]>('-a-|', { a: [() => of(null)] }), FIXTURE_PARAMS)
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        test('returns result from provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getCodeActions(
                        cold<ProvideCodeActionsSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_CODE_ACTIONS)],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_CODE_ACTIONS,
                })
            ))
    })

    test('errors do not propagate', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                getCodeActions(
                    cold<ProvideCodeActionsSignature[]>('-a-|', {
                        a: [() => of(FIXTURE_CODE_ACTIONS), () => throwError(new Error('x'))],
                    }),
                    FIXTURE_PARAMS,
                    false
                )
            ).toBe('-a-|', {
                a: FIXTURE_CODE_ACTIONS,
            })
        ))

    describe('2 providers', () => {
        test('returns null result if both providers return null', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getCodeActions(
                        cold<ProvideCodeActionsSignature[]>('-a-|', {
                            a: [() => of(null), () => of(null)],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: null,
                })
            ))

        test('omits null result from 1 provider', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getCodeActions(
                        cold<ProvideCodeActionsSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_CODE_ACTIONS), () => of(null)],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: FIXTURE_CODE_ACTIONS,
                })
            ))

        test('merges results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getCodeActions(
                        cold<ProvideCodeActionsSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_CODE_ACTIONS), () => of(FIXTURE_CODE_ACTIONS)],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-|', {
                    a: [...FIXTURE_CODE_ACTIONS, ...FIXTURE_CODE_ACTIONS],
                })
            ))
    })

    describe('multiple emissions', () => {
        test('returns stream of results', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getCodeActions(
                        cold<ProvideCodeActionsSignature[]>('-a-b-|', {
                            a: [() => of(FIXTURE_CODE_ACTIONS)],
                            b: [() => of(null)],
                        }),
                        FIXTURE_PARAMS
                    )
                ).toBe('-a-b-|', {
                    a: FIXTURE_CODE_ACTIONS,
                    b: null,
                })
            ))
    })
})
