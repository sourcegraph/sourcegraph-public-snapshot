import { ProxyMarked, proxyMarker, Remote } from 'comlink'
import { Observer } from 'rxjs'
import { FileDecorationsByPath } from 'sourcegraph'
import { SettingsCascade } from '../../settings/settings'
import { MainThreadAPI } from '../contract'
import { pretendRemote } from '../util'
import { ExtensionDocuments } from './api/documents'
import { initNewExtensionAPI } from './flatExtensionApi'

describe('getFileDecorations from ExtensionHost API, it aims to have more e2e feel', () => {
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

    it('restarts hover call if a provider was added or removed', () => {
        const documents = new ExtensionDocuments(() => Promise.resolve())

        const { exposedToMain, registerFileDecorationProvider } = initNewExtensionAPI(
            noopMain,
            emptySettings,
            documents
        )

        let counter = 0
        registerFileDecorationProvider({
            provideFileDecorations: ({ files }) => [{ path: files[0]?.path, text: { value: `a${++counter}` } }],
        })

        let results: FileDecorationsByPath[] = []
        exposedToMain
            .getFileDecorations({
                uri: 'git://gitlab.sgdev.org/sourcegraph/code-intel-extensions?1111#extensions/typescript/src/',
                files: [
                    {
                        uri:
                            'git://gitlab.sgdev.org/sourcegraph/code-intel-extensions?1111#extensions/typescript/src/package.ts',
                        isDirectory: false,
                        path: 'typescript/src/package.ts',
                    },
                ],
            })
            .subscribe(observe(value => results.push(value)))

        // first provider results
        expect(results).toEqual<FileDecorationsByPath[]>([
            {},
            {
                'typescript/src/package.ts': [
                    {
                        path: 'typescript/src/package.ts',
                        text: {
                            value: 'a1',
                        },
                    },
                ],
            },
        ])
        results = []

        const subscription = registerFileDecorationProvider({
            provideFileDecorations: ({ files }) => [{ path: files[0]?.path, text: { value: 'b' } }],
        })

        // second and first
        expect(results).toEqual<FileDecorationsByPath[]>([
            {
                'typescript/src/package.ts': [
                    {
                        path: 'typescript/src/package.ts',
                        text: {
                            value: 'a2',
                        },
                    },
                ],
            },
            {
                'typescript/src/package.ts': [
                    {
                        path: 'typescript/src/package.ts',
                        text: {
                            value: 'a2',
                        },
                    },
                    {
                        path: 'typescript/src/package.ts',
                        text: {
                            value: 'b',
                        },
                    },
                ],
            },
        ])
        results = []

        subscription.unsubscribe()

        // just first was queried for the third time
        expect(results).toEqual<FileDecorationsByPath[]>([
            {},
            {
                'typescript/src/package.ts': [
                    {
                        path: 'typescript/src/package.ts',
                        text: {
                            value: 'a3',
                        },
                    },
                ],
            },
        ])
        results = []
    })
})
