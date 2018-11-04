import * as assert from 'assert'
import { map } from 'rxjs/operators'
import { ConfigurationUpdate } from '../client/controller'
import { assertToJSON } from '../extension/types/common.test'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Configuration (integration)', () => {
    it('is usable in synchronous activation functions', async () => {
        // Test that extensions can access configuration synchronously in their `activate` functions.
        const { extensionHost } = await integrationTestContext()
        assert.doesNotThrow(() => extensionHost.configuration.subscribe(() => void 0))
        // TODO(sqs): Make the `get` function usable synchronously so this test passes.
        //
        // assert.doesNotThrow(() => extensionHost.configuration.get())
    })

    describe('Configuration#get', () => {
        it('gets configuration after ready', async () => {
            const { extensionHost, ready } = await integrationTestContext()
            await ready
            assertToJSON(extensionHost.configuration.get(), { a: 1 })
            assert.deepStrictEqual(extensionHost.configuration.get().value, { a: 1 })
        })
    })

    describe('Configuration#update', () => {
        it('updates configuration', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()
            await ready

            const values = collectSubscribableValues(
                clientController.configurationUpdates.pipe(map(({ path, value }) => ({ path, value })))
            )

            await extensionHost.configuration.get().update('a', 2)
            await extensionHost.internal.sync()
            assert.deepStrictEqual(values, [{ path: ['a'], value: 2 }] as ConfigurationUpdate[])
            values.length = 0 // clear

            await extensionHost.configuration.get().update('a', 3)
            await extensionHost.internal.sync()
            assert.deepStrictEqual(values, [{ path: ['a'], value: 3 }] as ConfigurationUpdate[])
        })
    })

    describe('configuration.subscribe', () => {
        it('subscribes to changes', async () => {
            const { clientController, extensionHost, getEnvironment, ready } = await integrationTestContext()
            await ready

            let calls = 0
            extensionHost.configuration.subscribe(() => calls++)
            assert.strictEqual(calls, 1) // called initially

            const prevEnvironment = getEnvironment()
            clientController.setEnvironment({
                ...prevEnvironment,
                configuration: { merged: { a: 3 } },
            })
            await extensionHost.internal.sync()
            assert.strictEqual(calls, 2)
        })
    })
})
