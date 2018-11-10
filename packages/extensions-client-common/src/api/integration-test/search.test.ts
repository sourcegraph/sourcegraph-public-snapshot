import * as assert from 'assert'
import { take } from 'rxjs/operators'
import { integrationTestContext } from './helpers.test'

describe('search (integration)', () => {
    it('registers a query transformer', async () => {
        const { clientController, extensionHost, ready } = await integrationTestContext()

        // Register the provider and call it
        const unsubscribe = extensionHost.search.registerQueryTransformer({ transformQuery: () => 'bar' })
        await ready
        assert.deepStrictEqual(
            await clientController.registries.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise(),
            'bar'
        )

        // Unregister the provider and ensure it's removed.
        unsubscribe.unsubscribe()
        assert.deepStrictEqual(
            await clientController.registries.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise(),
            'foo'
        )
    })

    it('supports multiple query transformers', async () => {
        const { clientController, extensionHost, ready } = await integrationTestContext()

        // Register the provider and call it
        extensionHost.search.registerQueryTransformer({ transformQuery: q => `${q} bar` })
        extensionHost.search.registerQueryTransformer({ transformQuery: q => `${q} qux` })
        await ready
        assert.deepStrictEqual(
            await clientController.registries.queryTransformer
                .transformQuery('foo')
                .pipe(take(1))
                .toPromise(),
            'foo bar qux'
        )
    })
})
