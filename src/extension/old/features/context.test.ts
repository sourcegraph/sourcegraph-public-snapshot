import assert from 'assert'
import { ContextUpdateNotification, ContextUpdateParams } from '../../../protocol/context'
import { MockConnection } from '../../../protocol/jsonrpc2/test/mockConnection'
import { ExtensionContext } from '../api'
import { createExtContext } from './context'

describe('ExtContext', () => {
    function create(): { extContext: ExtensionContext; mockConnection: MockConnection } {
        const mockConnection = new MockConnection()
        const extContext = createExtContext(mockConnection)
        return { extContext, mockConnection }
    }

    describe('updateContext', () => {
        it('sends to the client', async () => {
            const { extContext, mockConnection } = create()
            extContext.updateContext({ a: 'b' })
            assert.deepStrictEqual(mockConnection.sentMessages, [
                {
                    method: ContextUpdateNotification.type.method,
                    params: { updates: { a: 'b' } } as ContextUpdateParams,
                },
            ])
        })
    })
})
