import assert from 'assert'
import { MockMessageConnection } from '../../jsonrpc2/test/mockMessageConnection'
import {
    ConfigurationCascade,
    ConfigurationUpdateParams,
    ConfigurationUpdateRequest,
    DidChangeConfigurationNotification,
    DidChangeConfigurationParams,
} from '../../protocol'
import { Configuration, Observable } from '../api'
import { observableValue } from '../util'
import { createExtConfiguration } from './configuration'

interface Settings {
    [key: string]: string
}

describe('ExtConfiguration', () => {
    function create<C = Settings>(
        initial: ConfigurationCascade<C> = { merged: {} as C }
    ): {
        extConfiguration: Configuration<C> & Observable<C>
        mockConnection: MockMessageConnection
    } {
        const mockConnection = new MockMessageConnection()
        const extConfiguration = createExtConfiguration({ rawConnection: mockConnection }, initial)
        return { extConfiguration, mockConnection }
    }

    it('starts with initial configuration', () => {
        const { extConfiguration } = create({ merged: { a: 'b' } })
        assert.deepStrictEqual(observableValue(extConfiguration), { a: 'b' } as Settings)
    })

    it('reflects client updates', () => {
        const { extConfiguration, mockConnection } = create({ merged: { a: 'b' } })
        mockConnection.recvNotification(DidChangeConfigurationNotification.type.method, {
            configurationCascade: { merged: { c: 'd' } as Settings },
        } as DidChangeConfigurationParams)
        assert.deepStrictEqual(observableValue(extConfiguration), { c: 'd' } as Settings)
    })

    describe('update', () => {
        it('sends to the client', async () => {
            const { extConfiguration, mockConnection } = create()
            mockConnection.mockResults.set(ConfigurationUpdateRequest.type.method, void 0)
            await extConfiguration.update('a', 'b')
            assert.deepStrictEqual(mockConnection.sentMessages, [
                {
                    method: ConfigurationUpdateRequest.type.method,
                    params: { path: ['a'], value: 'b' } as ConfigurationUpdateParams,
                },
            ])
        })
    })
})
