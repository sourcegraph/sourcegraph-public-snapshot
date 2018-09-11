import * as assert from 'assert'
import { ContextUpdateNotification, ContextUpdateParams } from '../../protocol/context'
import { NotificationHandler } from '../../protocol/jsonrpc2/handlers'
import { Client } from '../client'
import { ContextFeature } from './context'

describe('ContextFeature', () => {
    const create = (
        setContext?: (params: ContextUpdateParams) => void
    ): {
        client: Client
        feature: ContextFeature
    } => {
        const client = { options: {} } as Client
        const feature = new ContextFeature(
            client,
            setContext ||
                (() => {
                    throw new Error('no setContext')
                })
        )
        return { client, feature }
    }

    describe('upon initialization', () => {
        it('registers the provider and listens for notifications', done => {
            const { client, feature } = create()

            function mockOnNotification(method: string, _handler: NotificationHandler<any>): void {
                assert.strictEqual(method, ContextUpdateNotification.type)
                done()
            }
            client.onNotification = mockOnNotification

            feature.initialize()
        })
    })

    describe('on context/update', () => {
        it('updates context', () => {
            let setContextParams: ContextUpdateParams | undefined
            const { client, feature } = create(params => (setContextParams = params))

            const params: ContextUpdateParams = { updates: { a: 1, b: '2', c: true, d: null } }
            client.onNotification = (_type: string, handler: NotificationHandler<ContextUpdateParams>) => {
                handler(params)
            }

            feature.initialize()

            assert.deepStrictEqual(setContextParams, params)
        })
    })
})
