import { take } from 'rxjs/operators'
import { integrationTestContext } from './testHelpers'

describe('search (integration)', () => {
    test('registers a query transformer', async () => {
        const { services, extensionHost } = await integrationTestContext()

        // Register the provider and call it
        const unsubscribe = extensionHost.search.registerQueryTransformer({ transformQuery: () => 'bar' })
        await extensionHost.internal.sync()
        expect(
            await services.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual('bar')

        // Unregister the provider and ensure it's removed.
        unsubscribe.unsubscribe()
        expect(
            await services.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual('foo')
    })

    test('supports multiple query transformers', async () => {
        const { services, extensionHost } = await integrationTestContext()

        // Register the provider and call it
        extensionHost.search.registerQueryTransformer({ transformQuery: (q: string) => `${q} bar` })
        extensionHost.search.registerQueryTransformer({ transformQuery: (q: string) => `${q} qux` })
        await extensionHost.internal.sync()
        expect(
            await services.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual('foo bar qux')
    })
})
