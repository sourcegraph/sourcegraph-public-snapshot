import * as assert from 'assert'
import { ContextValues } from 'sourcegraph'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Context (integration)', () => {
    describe('internal.updateContext', () => {
        it('updates context', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()
            await ready
            const values = collectSubscribableValues(clientController.environment.context)

            extensionHost.internal.updateContext({ a: 1 })
            await extensionHost.internal.sync()
            assert.deepStrictEqual(values, [{}, { a: 1 }] as ContextValues[])
        })
    })
})
