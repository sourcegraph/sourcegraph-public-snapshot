import { from } from 'rxjs'
import { first, take } from 'rxjs/operators'
import { TextSearchResult } from 'sourcegraph'
import { integrationTestContext } from './testHelpers'

describe('search (integration)', () => {
    test('registers a query transformer', async () => {
        const { services, extensionAPI } = await integrationTestContext()

        // Register the provider and call it
        extensionAPI.search.registerQueryTransformer({
            transformQuery: () => 'bar',
        })
        await extensionAPI.internal.sync()
        expect(
            await services.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual('bar')
    })

    test('unregisters a query transformer', async () => {
        const { services, extensionAPI } = await integrationTestContext()

        // Register the provider and call it
        const subscription = extensionAPI.search.registerQueryTransformer({
            transformQuery: () => 'bar',
        })
        await extensionAPI.internal.sync()
        // Unregister the provider and ensure it's removed.
        subscription.unsubscribe()
        await extensionAPI.internal.sync()
        expect(
            await services.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual('foo')
    })

    test('supports multiple query transformers', async () => {
        const { services, extensionAPI } = await integrationTestContext()

        // Register the provider and call it
        extensionAPI.search.registerQueryTransformer({ transformQuery: (q: string) => `${q} bar` })
        extensionAPI.search.registerQueryTransformer({ transformQuery: (q: string) => `${q} qux` })
        await extensionAPI.internal.sync()
        expect(
            await services.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual('foo bar qux')
    })

    describe('findTextInFiles', () => {
        test('', async () => {
            const { extensionAPI } = await integrationTestContext()
            const results = await from(extensionAPI.search.findTextInFiles({ pattern: 'p', type: 'regexp' }))
                .pipe(first())
                .toPromise()
            expect(results.length).toBe(1)
            expect(results[0].uri.toString()).toBe('file:///SR1')
        })
    })
})
