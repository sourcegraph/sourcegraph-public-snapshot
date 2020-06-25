import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { TestScheduler } from 'rxjs/testing'
import { Hover } from 'sourcegraph'
import { getHover, ProvideTextDocumentHoverSignature } from './hover'
import { FIXTURE } from '../client/services/registry.test'
import { HoverMerged } from '../client/types/hover'
import { initNewExtensionAPI } from './flatExtensionApi'
import { pretendRemote } from '../util'
import { MainThreadAPI } from '../contract'
import { SettingsCascade } from '../../settings/settings'
import { Observer, Subject } from 'rxjs'
import { ProxyMarked, proxyMarker, Remote } from 'comlink'
import { ExtensionDocuments } from './api/documents'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

const FIXTURE_RESULT: Hover | null = { contents: { value: 'c', kind: MarkupKind.PlainText } }
const FIXTURE_RESULT_MERGED: HoverMerged | null = { contents: [{ value: 'c', kind: MarkupKind.PlainText }] }

describe('getHover standalone function', () => {
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

describe('getHover from ExtensionHost API', () => {
    const noopMain = pretendRemote<MainThreadAPI>({})
    const emptySettings: SettingsCascade<object> = { subjects: [], final: {} }

    const observe = <T>(onValue: (val: T) => void): Remote<Observer<T> & ProxyMarked> =>
        pretendRemote({
            next: onValue,
            error: (error: any) => {
                throw error
            },
            complete: () => {},
            [proxyMarker]: Promise.resolve(true as const),
        })

    const textHover = (value: string): Hover => ({ contents: { value, kind: MarkupKind.PlainText } })

    describe('integration(ish) tests for scenarios not covered by tests above', () => {
        it('it filters hover providers', () => {
            const typescriptFileUri = 'file:///f.ts'
            const documents = new ExtensionDocuments(() => Promise.resolve())
            documents.$acceptDocumentData([{ type: 'added', languageId: 'ts', text: 'body', uri: typescriptFileUri }])

            const { exposedToMain, languages } = initNewExtensionAPI(noopMain, emptySettings, documents)

            languages.registerHoverProvider([{ pattern: '*.js' }], {
                provideHover: () => textHover('js'),
            })

            const tsHover: Hover = textHover('ts')
            languages.registerHoverProvider([{ pattern: '*.ts' }], {
                provideHover: () => tsHover,
            })

            const results: any[] = []
            exposedToMain
                .getHover({ position: { line: 1, character: 2 }, textDocument: { uri: typescriptFileUri } })
                .subscribe(observe(value => results.push(value)))

            expect(results).toEqual<MaybeLoadingResult<HoverMerged | null>[]>([
                { isLoading: true, result: null },
                { isLoading: false, result: { contents: [tsHover.contents] } },
            ])
        })

        it('adds hover results if a provider was added in the middle of the execution', () => {
            const typescriptFileUri = 'file:///f.ts'
            const documents = new ExtensionDocuments(() => Promise.resolve())
            documents.$acceptDocumentData([{ type: 'added', languageId: 'ts', text: 'body', uri: typescriptFileUri }])

            const { exposedToMain, languages } = initNewExtensionAPI(noopMain, emptySettings, documents)

            const hovers = new Subject<Hover>()
            // let calledTimes = 0

            languages.registerHoverProvider([{ pattern: '*.ts' }], {
                provideHover: () => {
                    // calledTimes++
                    return hovers
                },
            })
            const results: any[] = []
            exposedToMain
                .getHover({ position: { line: 1, character: 2 }, textDocument: { uri: typescriptFileUri } })
                .subscribe(observe(value => results.push(value)))

            hovers.next(textHover('a'))

            languages.registerHoverProvider([{ pattern: '*.ts' }], {
                provideHover: () => textHover('b'),
            })

            hovers.next(textHover('c'))
            hovers.complete()

            expect(results).toEqual<MaybeLoadingResult<HoverMerged | null>[]>([
                { isLoading: true, result: null },
                // 'a' from the first provider
                // TODO(simon) shouldn't there be a loading 'true' state?
                { isLoading: false, result: { contents: [textHover('a').contents] } },
                // 'c' -> first,  'b'-> second
                { isLoading: false, result: { contents: ['c', 'b'].map(value => textHover(value).contents) } },
            ])
        })

        it('restarts hover call if a provider was added', () => {
            const typescriptFileUri = 'file:///f.ts'
            const documents = new ExtensionDocuments(() => Promise.resolve())
            documents.$acceptDocumentData([{ type: 'added', languageId: 'ts', text: 'body', uri: typescriptFileUri }])

            const { exposedToMain, languages } = initNewExtensionAPI(noopMain, emptySettings, documents)

            let counter = 0
            languages.registerHoverProvider([{ pattern: '*.ts' }], {
                provideHover: () => textHover('a' + ++counter),
            })

            const results: any[] = []
            exposedToMain
                .getHover({ position: { line: 1, character: 2 }, textDocument: { uri: typescriptFileUri } })
                .subscribe(observe(value => results.push(value)))

            languages.registerHoverProvider([{ pattern: '*.ts' }], {
                provideHover: () => textHover('b'),
            })

            expect(results).toEqual<MaybeLoadingResult<HoverMerged | null>[]>([
                { isLoading: true, result: null },
                { isLoading: false, result: { contents: [textHover('a1').contents] } },
                // TODO(simon) do we actually need loading here?
                { isLoading: true, result: { contents: [textHover('a2').contents] } },
                { isLoading: false, result: { contents: ['a2', 'b'].map(value => textHover(value).contents) } },
            ])
        })

        it('restarts hover query if a provider was deleted', () => {
            const typescriptFileUri = 'file:///f.ts'
            const documents = new ExtensionDocuments(() => Promise.resolve())
            documents.$acceptDocumentData([{ type: 'added', languageId: 'ts', text: 'body', uri: typescriptFileUri }])

            const { exposedToMain, languages } = initNewExtensionAPI(noopMain, emptySettings, documents)

            let counter = 0
            languages.registerHoverProvider([{ pattern: '*.ts' }], {
                provideHover: () => textHover('a' + ++counter),
            })

            const subscription = languages.registerHoverProvider([{ pattern: '*.ts' }], {
                provideHover: () => textHover('b'),
            })

            const results: any[] = []
            exposedToMain
                .getHover({ position: { line: 1, character: 2 }, textDocument: { uri: typescriptFileUri } })
                .subscribe(observe(value => results.push(value)))

            subscription.unsubscribe()

            expect(results).toEqual<MaybeLoadingResult<HoverMerged | null>[]>([
                { isLoading: true, result: { contents: [textHover('a1').contents] } },
                { isLoading: false, result: { contents: ['a1', 'b'].map(value => textHover(value).contents) } },
                { isLoading: true, result: null },
                { isLoading: false, result: { contents: [textHover('a2').contents] } },
            ])
        })
    })

    //
})
