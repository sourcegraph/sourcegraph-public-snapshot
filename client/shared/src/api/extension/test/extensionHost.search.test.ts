import { ProxyMarked, proxyMarker, Remote } from 'comlink'
import { BehaviorSubject, Observer } from 'rxjs'

import { SettingsCascade } from '../../../settings/settings'
import { ClientAPI } from '../../client/api/api'
import { pretendRemote } from '../../util'
import { proxySubscribable } from '../api/common'

import { initializeExtensionHostTest } from './test-helpers'

const noopMain = pretendRemote<ClientAPI>({
    getEnabledExtensions: () => proxySubscribable(new BehaviorSubject([])),
})
const initialSettings: SettingsCascade<object> = { subjects: [], final: {} }

const observe = (onValue: (value: string) => void): Remote<Observer<string> & ProxyMarked> =>
    pretendRemote({
        next: onValue,
        error: (error: any) => {
            throw error
        },
        complete: () => {},
        [proxyMarker]: Promise.resolve(true as const),
    })

describe('QueryTransformers', () => {
    it('returns the same query with no registered transformers', () => {
        const { extensionHostAPI } = initializeExtensionHostTest(
            { initialSettings, clientApplication: 'sourcegraph', sourcegraphURL: 'https://example.com/' },
            noopMain
        )

        const results: string[] = []
        extensionHostAPI.transformSearchQuery('a').subscribe(observe(value => results.push(value)))
        expect(results).toEqual(['a'])
    })

    it('can work with Promise based transformers', async () => {
        const { extensionHostAPI, extensionAPI } = initializeExtensionHostTest(
            { initialSettings, clientApplication: 'sourcegraph', sourcegraphURL: 'https://example.com/' },
            noopMain
        )
        extensionAPI.search.registerQueryTransformer({ transformQuery: query => Promise.resolve(query + '!') })
        const result = await new Promise(resolve =>
            extensionHostAPI.transformSearchQuery('a').subscribe(observe(resolve))
        )
        expect(result).toEqual('a!')
    })

    it('emits a new transformed value if there is a new transformer', () => {
        const { extensionHostAPI, extensionAPI } = initializeExtensionHostTest(
            { initialSettings, clientApplication: 'sourcegraph', sourcegraphURL: 'https://example.com/' },
            noopMain
        )

        const results: string[] = []
        extensionHostAPI.transformSearchQuery('a').subscribe(observe(value => results.push(value)))
        expect(results).toEqual(['a'])

        extensionAPI.search.registerQueryTransformer({ transformQuery: query => query + '!' })
        expect(results).toEqual(['a', 'a!'])
    })

    it('emits new value if a transformer was removed', () => {
        const { extensionHostAPI, extensionAPI } = initializeExtensionHostTest(
            { initialSettings, clientApplication: 'sourcegraph', sourcegraphURL: 'https://example.com/' },
            noopMain
        )

        const transformerSubscription = extensionAPI.search.registerQueryTransformer({
            transformQuery: query => query + '!',
        })

        const results: string[] = []
        extensionHostAPI.transformSearchQuery('a').subscribe(observe(value => results.push(value)))
        expect(results).toEqual(['a!'])
        transformerSubscription.unsubscribe()

        expect(results).toEqual(['a!', 'a'])
    })

    it('emits modified query if there are any transformers registered', () => {
        const { extensionHostAPI, extensionAPI } = initializeExtensionHostTest(
            { initialSettings, clientApplication: 'sourcegraph', sourcegraphURL: 'https://example.com/' },
            noopMain
        )

        extensionAPI.search.registerQueryTransformer({ transformQuery: query => query + '!' })

        const results: string[] = []
        extensionHostAPI.transformSearchQuery('a').subscribe(observe(value => results.push(value)))
        expect(results).toEqual(['a!'])
    })

    it('cancels previous transformer chains', async () => {
        const { extensionHostAPI, extensionAPI } = initializeExtensionHostTest(
            { initialSettings, clientApplication: 'sourcegraph', sourcegraphURL: 'https://example.com/' },
            noopMain
        )

        // collect all pending promises and their triggers from the first transformer
        // to manually manipulate them later
        const resolves: ((q: string) => void)[] = []
        const promises: Promise<string>[] = []

        extensionAPI.search.registerQueryTransformer({
            transformQuery: () => {
                const promise = new Promise<string>(resolve => {
                    resolves.push(resolve)
                })
                promises.push(promise)
                return promise
            },
        })

        let secondTransformerCallCount = 0
        extensionAPI.search.registerQueryTransformer({
            transformQuery: query => {
                secondTransformerCallCount++
                return query
            },
        })

        const results: string[] = []

        extensionHostAPI.transformSearchQuery('a').subscribe(observe(value => results.push(value)))
        // we hanged in the first transformer
        expect(results).toEqual([])
        expect(promises.length).toBe(1)

        extensionAPI.search.registerQueryTransformer({ transformQuery: query => query + '!' })

        // we reissued the transformation and waiting for a new promise to resolve
        expect(promises.length).toBe(2)
        // means that we haven't executed the second transformer yet
        // because it scheduled after the promise
        expect(secondTransformerCallCount).toBe(0)

        for (const resolve of resolves) {
            resolve('finally')
        }

        await Promise.all(promises)

        // the first chain was aborted
        expect(secondTransformerCallCount).toBe(1)
        expect(results).toEqual(['finally!'])
    })
})
