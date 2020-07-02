import { Range, DocumentHighlight } from 'sourcegraph'
import { initNewExtensionAPI, mergeDocumentHighlightResults } from './flatExtensionApi'
import { pretendRemote } from '../util'
import { MainThreadAPI } from '../contract'
import { SettingsCascade } from '../../settings/settings'
import { Observer } from 'rxjs'
import { ProxyMarked, proxyMarker, Remote } from 'comlink'
import { ExtensionDocuments } from './api/documents'
import { MaybeLoadingResult, LOADING } from '@sourcegraph/codeintellify'

// TODO this is weird to cast to ranges
const range1 = ({ start: { line: 1, character: 2 }, end: { line: 3, character: 4 } } as unknown) as Range
const range2 = ({ start: { line: 2, character: 3 }, end: { line: 4, character: 5 } } as unknown) as Range
const range3 = ({ start: { line: 3, character: 4 }, end: { line: 5, character: 6 } } as unknown) as Range

describe('mergeDocumentHighlightResults', () => {
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

describe('getDocumentHighlights from ExtensionHost API, it aims to have more e2e feel', () => {
    // integration(ish) tests for scenarios not covered by providers tests
    const noopMain = pretendRemote<MainThreadAPI>({})
    const emptySettings: SettingsCascade<object> = {
        subjects: [],
        final: {},
    }

    const observe = <T>(onValue: (val: T) => void): Remote<Observer<T> & ProxyMarked> =>
        pretendRemote({
            next: onValue,
            error: (error: any) => {
                throw error
            },
            complete: () => {},
            [proxyMarker]: Promise.resolve(true as const),
        })

    const documentHighlight = (value: number): DocumentHighlight => ({
        // TODO this is weird to cast to ranges
        range: ({
            start: { line: value, character: value },
            end: { line: value, character: value },
        } as unknown) as Range,
    })

    it('restarts document highlights call if a provider was added or removed', () => {
        const typescriptFileUri = 'file:///f.ts'
        const documents = new ExtensionDocuments(() => Promise.resolve())
        documents.$acceptDocumentData([
            {
                type: 'added',
                languageId: 'ts',
                text: 'body',
                uri: typescriptFileUri,
            },
        ])

        const { exposedToMain, languages } = initNewExtensionAPI(noopMain, emptySettings, documents)

        let counter = 0
        languages.registerDocumentHighlightProvider([{ pattern: '*.ts' }], {
            provideDocumentHighlights: () => [documentHighlight(++counter)],
        })

        let results: any[] = []
        exposedToMain
            .getDocumentHighlights({
                position: { line: 1, character: 2 },
                textDocument: { uri: typescriptFileUri },
            })
            .subscribe(observe(value => results.push(value)))

        // first provider results
        expect(results).toEqual<MaybeLoadingResult<DocumentHighlight[] | null>[]>([
            { isLoading: true, result: [] },
            {
                isLoading: false,
                result: [documentHighlight(1)],
            },
        ])
        results = []

        const subscription = languages.registerDocumentHighlightProvider([{ pattern: '*.ts' }], {
            provideDocumentHighlights: () => [documentHighlight(-1)],
        })

        // second and first
        expect(results).toEqual<MaybeLoadingResult<DocumentHighlight[] | null>[]>([
            {
                isLoading: true,
                result: [documentHighlight(2)],
            },
            {
                isLoading: false,
                result: [2, -1].map(value => documentHighlight(value)),
            },
        ])
        results = []

        subscription.unsubscribe()

        // just first was queried for the third time
        expect(results).toEqual<MaybeLoadingResult<DocumentHighlight[] | null>[]>([
            { isLoading: true, result: [] },
            {
                isLoading: false,
                result: [documentHighlight(3)],
            },
        ])
    })
})
