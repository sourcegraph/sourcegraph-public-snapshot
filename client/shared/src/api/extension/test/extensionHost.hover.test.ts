import { describe, expect, it } from '@jest/globals'
import { type ProxyMarked, proxyMarker, type Remote } from 'comlink'
import { type Observer, of } from 'rxjs'
import type { Hover } from 'sourcegraph'

import type { HoverMerged } from '@sourcegraph/client-api'
import type { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { MarkupKind } from '@sourcegraph/extension-api-classes'

import type { SettingsCascade } from '../../../settings/settings'
import type { ClientAPI } from '../../client/api/api'
import { pretendProxySubscribable, pretendRemote } from '../../util'

import { initializeExtensionHostTest } from './test-helpers'

describe('getHover from ExtensionHost API, it aims to have more e2e feel', () => {
    // integration(ish) tests for scenarios not covered by providers tests
    const noopMain = pretendRemote<ClientAPI>({
        getEnabledExtensions: () => pretendProxySubscribable(of([])),
    })
    const initialSettings: SettingsCascade<object> = {
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

    const textHover = (value: string): Hover => ({
        contents: { value, kind: MarkupKind.PlainText },
    })

    it('restarts hover call if a provider was added or removed', () => {
        const typescriptFileUri = 'file:///f.ts'

        const { extensionHostAPI, extensionAPI } = initializeExtensionHostTest(
            { initialSettings, clientApplication: 'sourcegraph', sourcegraphURL: 'https://example.com/' },
            noopMain
        )

        extensionHostAPI.addTextDocumentIfNotExists({
            languageId: 'ts',
            text: 'body',
            uri: typescriptFileUri,
        })

        let counter = 0
        extensionAPI.languages.registerHoverProvider([{ pattern: '*.ts' }], {
            provideHover: () => textHover(`a${++counter}`),
        })

        let results: any[] = []
        extensionHostAPI
            .getHover({
                position: { line: 1, character: 2 },
                textDocument: { uri: typescriptFileUri },
            })
            .subscribe(observe(value => results.push(value)))

        // first provider results
        expect(results).toEqual([
            { isLoading: true, result: null },
            {
                isLoading: false,
                result: { contents: [textHover('a1').contents], aggregatedBadges: [] },
            },
        ] as MaybeLoadingResult<HoverMerged | null>[])
        results = []

        const subscription = extensionAPI.languages.registerHoverProvider([{ pattern: '*.ts' }], {
            provideHover: () => textHover('b'),
        })

        // second and first
        expect(results).toEqual([
            {
                isLoading: true,
                result: { contents: [textHover('a2').contents], aggregatedBadges: [] },
            },
            {
                isLoading: false,
                result: {
                    contents: ['a2', 'b'].map(value => textHover(value).contents),
                    aggregatedBadges: [],
                },
            },
        ] as MaybeLoadingResult<HoverMerged | null>[])
        results = []

        subscription.unsubscribe()

        // just first was queried for the third time
        expect(results).toEqual([
            { isLoading: true, result: null },
            {
                isLoading: false,
                result: { contents: [textHover('a3').contents], aggregatedBadges: [] },
            },
        ] as MaybeLoadingResult<HoverMerged | null>[])
    })
})
