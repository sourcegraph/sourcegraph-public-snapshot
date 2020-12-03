import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { of } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { TextDocumentIdentifier } from '../types/textDocument'
import {
    decorationAttachmentStyleForTheme,
    decorationStyleForTheme,
    fileDecorationColorForTheme,
    getDecorations,
    ProvideTextDocumentDecorationSignature,
} from './decoration'
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

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('getDecorations', () => {
    describe('0 providers', () => {
        test('returns null', () =>
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
        test('returns null result from provider', () =>
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

        test('returns result from provider', () =>
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
        test('returns null result if both providers return null', () =>
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

        test('omits null result from 1 provider', () =>
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

        test('merges results from providers', () =>
            scheduler().run(({ cold, expectObservable }) =>
                expectObservable(
                    getDecorations(
                        cold<ProvideTextDocumentDecorationSignature[]>('-a-|', {
                            a: [() => of(FIXTURE_RESULT), () => of(FIXTURE_RESULT)],
                        }),
                        FIXTURE.TextDocumentIdentifier
                    )
                ).toBe('-a-|', {
                    a: [...FIXTURE_RESULT, ...FIXTURE_RESULT],
                })
            ))
    })

    describe('multiple emissions', () => {
        test('returns stream of results', () =>
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

describe('decorationStyleForTheme', () => {
    const FIXTURE_RANGE = { start: { line: 1, character: 2 }, end: { line: 3, character: 4 } }

    test('supports no theme overrides', () =>
        expect(decorationStyleForTheme({ range: FIXTURE_RANGE, backgroundColor: 'red' }, true)).toEqual({
            backgroundColor: 'red',
        }))

    test('applies light theme overrides', () =>
        expect(
            decorationStyleForTheme(
                { range: FIXTURE_RANGE, backgroundColor: 'red', light: { backgroundColor: 'blue' } },
                true
            )
        ).toEqual({
            backgroundColor: 'blue',
        }))

    test('applies dark theme overrides', () =>
        expect(
            decorationStyleForTheme(
                {
                    range: FIXTURE_RANGE,
                    backgroundColor: 'red',
                    light: { backgroundColor: 'blue' },
                    dark: { backgroundColor: 'green' },
                },
                false
            )
        ).toEqual({
            backgroundColor: 'green',
        }))
})

describe('decorationAttachmentStyleForTheme', () => {
    test('supports no theme overrides', () =>
        expect(decorationAttachmentStyleForTheme({ color: 'red' }, true)).toEqual({ color: 'red' }))

    test('applies light theme overrides', () =>
        expect(decorationAttachmentStyleForTheme({ color: 'red', light: { color: 'blue' } }, true)).toEqual({
            color: 'blue',
        }))

    test('applies dark theme overrides', () =>
        expect(
            decorationAttachmentStyleForTheme(
                { color: 'red', light: { color: 'blue' }, dark: { color: 'green' } },
                false
            )
        ).toEqual({
            color: 'green',
        }))
})

describe('fileDecorationColorForTheme', () => {
    test('supports no theme overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                },
                false
            )
        ).toEqual('red')
    })

    test('applies light theme overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    light: {
                        color: 'blue',
                    },
                },
                true
            )
        ).toEqual('blue')
    })

    test('applies dark theme overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    dark: {
                        color: 'green',
                    },
                },
                false
            )
        ).toEqual('green')
    })

    test('applies selected color overrides', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    activeColor: 'orange',
                },
                false,
                true
            )
        ).toEqual('orange')
    })

    test('applies selected color for themes', () => {
        expect(
            fileDecorationColorForTheme(
                {
                    contentText: '',
                    color: 'red',
                    activeColor: 'orange',
                    dark: {
                        activeColor: 'teal',
                    },
                },
                false,
                true
            )
        ).toEqual('teal')
    })
})
