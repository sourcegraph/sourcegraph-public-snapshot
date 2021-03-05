import { DocumentHighlight } from 'sourcegraph'
import { Range } from '@sourcegraph/extension-api-classes'
import { initNewExtensionAPI } from './flatExtensionApi'
import { pretendRemote } from '../util'
import { MainThreadAPI } from '../contract'
import { SettingsCascade } from '../../settings/settings'
import { Observer } from 'rxjs'
import { ProxyMarked, proxyMarker, Remote } from 'comlink'
import { ExtensionDocuments } from './api/documents'

describe('getDocumentHighlights from ExtensionHost API, it aims to have more e2e feel', () => {
    // integration(ish) tests for scenarios not covered by providers tests
    const noopMain = pretendRemote<MainThreadAPI>({})
    const emptySettings: SettingsCascade<object> = {
        subjects: [],
        final: {},
    }

    const observe = <T>(onValue: (value: T) => void): Remote<Observer<T> & ProxyMarked> =>
        pretendRemote({
            next: onValue,
            error: (error: any) => {
                throw error
            },
            complete: () => {},
            [proxyMarker]: Promise.resolve(true as const),
        })

    const documentHighlight = (value: number): DocumentHighlight => ({
        range: new Range(value, value, value, value),
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

        let results: DocumentHighlight[][] = []
        exposedToMain
            .getDocumentHighlights({
                position: { line: 1, character: 2 },
                textDocument: { uri: typescriptFileUri },
            })
            .subscribe(observe(value => results.push(value)))

        // first provider results
        expect(results).toEqual([[], [documentHighlight(1)]])
        results = []

        const subscription = languages.registerDocumentHighlightProvider([{ pattern: '*.ts' }], {
            provideDocumentHighlights: () => [documentHighlight(0)],
        })

        // second and first
        expect(results).toEqual([[], [2, 0].map(value => documentHighlight(value))])
        results = []

        subscription.unsubscribe()

        // just first was queried for the third time
        expect(results).toEqual([[], [documentHighlight(3)]])
    })
})
