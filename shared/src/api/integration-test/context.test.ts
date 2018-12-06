import * as assert from 'assert'
import { from } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'
import { ContextValues } from 'sourcegraph'
import { collectSubscribableValues, integrationTestContext } from './helpers.test'

describe('Context (integration)', () => {
    describe('internal.updateContext', () => {
        it('updates context', async () => {
            const { services, extensionHost } = await integrationTestContext()
            const values = collectSubscribableValues(from(services.context.data).pipe(distinctUntilChanged()))

            extensionHost.internal.updateContext({ a: 1 })
            await extensionHost.internal.sync()
            assert.deepStrictEqual(values, [
                { 'clientApplication.isSourcegraph': true, 'clientApplication.extensionAPIVersion.major': 3 },
                { a: 1, 'clientApplication.isSourcegraph': true, 'clientApplication.extensionAPIVersion.major': 3 },
            ] as ContextValues[])
        })
    })
})
