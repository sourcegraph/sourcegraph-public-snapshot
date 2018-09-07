import * as assert from 'assert'
import { NotificationHandler } from '../../jsonrpc2/handlers'
import { NotificationType } from '../../jsonrpc2/messages'
import { ContextUpdateNotification, ContextUpdateParams } from '../../protocol/context'
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

            function mockOnNotification(method: string, handler: NotificationHandler<any>): void
            function mockOnNotification(
                type: NotificationType<ContextUpdateParams, void>,
                params: NotificationHandler<ContextUpdateParams>
            ): void
            function mockOnNotification(
                type: string | NotificationType<ContextUpdateParams, void>,
                params: NotificationHandler<any>
            ): void {
                assert.strictEqual(typeof type === 'string' ? type : type.method, ContextUpdateNotification.type.method)
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
            client.onNotification = (
                _type: NotificationType<ContextUpdateParams, void> | string,
                handler: NotificationHandler<ContextUpdateParams>
            ) => {
                handler(params)
            }

            feature.initialize()

            assert.deepStrictEqual(setContextParams, params)
        })
    })
})
