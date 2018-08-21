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
import { createExtConfiguration, setValueAtKeyPath } from './configuration'

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
        it('sends to the client and immediately reflects locally', async () => {
            const { extConfiguration, mockConnection } = create()
            mockConnection.mockResults.set(ConfigurationUpdateRequest.type.method, void 0)
            const updated = extConfiguration.update('a', 'b')
            const want = { a: 'b' } as Settings
            assert.deepStrictEqual(observableValue(extConfiguration), want)
            await updated
            assert.deepStrictEqual(observableValue(extConfiguration), want)
            assert.deepStrictEqual(mockConnection.sentMessages, [
                {
                    method: ConfigurationUpdateRequest.type.method,
                    params: { path: ['a'], value: 'b' } as ConfigurationUpdateParams,
                },
            ])
        })

        it('handles interleaved update calls and didChangeConfiguration notifications', async () => {
            const { extConfiguration, mockConnection } = create()
            mockConnection.mockResults.set(ConfigurationUpdateRequest.type.method, void 0)
            const updated = extConfiguration.update('a', 'b')
            assert.deepStrictEqual(observableValue(extConfiguration), { a: 'b' } as Settings)
            mockConnection.recvNotification(DidChangeConfigurationNotification.type.method, {
                configurationCascade: { merged: { c: 'd' } as Settings },
            } as DidChangeConfigurationParams)
            assert.deepStrictEqual(observableValue(extConfiguration), { c: 'd' } as Settings)
            await updated
            assert.deepStrictEqual(observableValue(extConfiguration), { c: 'd' } as Settings)
        })
    })
})

describe('setValueAtKeyPath', () => {
    it('overwrites the top level', () => assert.deepStrictEqual(setValueAtKeyPath({ a: 1 }, [], { b: 2 }), { b: 2 }))
    it('overwrites an existing property', () => assert.deepStrictEqual(setValueAtKeyPath({ a: 1 }, ['a'], 2), { a: 2 }))
    it('sets a new property', () => assert.deepStrictEqual(setValueAtKeyPath({ a: 1 }, ['b'], 2), { a: 1, b: 2 }))
    it('sets a property overwriting an array', () => assert.deepStrictEqual(setValueAtKeyPath([1], ['a'], 2), { a: 2 }))
    it('sets a property overwriting a primitive', () =>
        assert.deepStrictEqual(setValueAtKeyPath(1 as any, ['a'], 2), { a: 2 }))
    it('overwrites an existing nested property', () =>
        assert.deepStrictEqual(setValueAtKeyPath({ a: { b: 1 } }, ['a', 'b'], 2), { a: { b: 2 } }))
    it('deletes a property', () =>
        assert.deepStrictEqual(setValueAtKeyPath({ a: 1, b: 2 }, ['a'], undefined), { b: 2 }))
    it('sets a new nested property', () =>
        assert.deepStrictEqual(setValueAtKeyPath({ a: { b: 1 } }, ['a', 'c'], 2), { a: { b: 1, c: 2 } }))
    it('sets a new deeply nested property', () =>
        assert.deepStrictEqual(setValueAtKeyPath({ a: { b: { c: 1 } } }, ['a', 'b', 'd'], 2), {
            a: { b: { c: 1, d: 2 } },
        }))
    it('overwrites an object', () => assert.deepStrictEqual(setValueAtKeyPath({ a: { b: 1 } }, ['a'], 2), { a: 2 }))
    it('sets a value that requires a new object', () =>
        assert.deepStrictEqual(setValueAtKeyPath({}, ['a', 'b'], 1), { a: { b: 1 } }))

    it('overwrites an existing index', () => assert.deepStrictEqual(setValueAtKeyPath([1], [0], 2), [2]))
    it('inserts a new index', () => assert.deepStrictEqual(setValueAtKeyPath([1], [1], 2), [1, 2]))
    it('inserts a new index at end', () => assert.deepStrictEqual(setValueAtKeyPath([1, 2], [-1], 3), [1, 2, 3]))
    it('inserts an index overwriting an object', () => assert.deepStrictEqual(setValueAtKeyPath({ a: 1 }, [0], 2), [2]))
    it('inserts an index overwriting a primitive', () =>
        assert.deepStrictEqual(setValueAtKeyPath(1 as any, [0], 2), [2]))
    it('overwrites an existing nested index', () =>
        assert.deepStrictEqual(setValueAtKeyPath([1, [2]], [1, 0], 3), [1, [3]]))
    it('deletes an index', () => assert.deepStrictEqual(setValueAtKeyPath([1, 2, 3], [1], undefined), [1, 3]))
    it('sets a new nested index', () =>
        assert.deepStrictEqual(setValueAtKeyPath([1, [1, 2, [1, 2, 3, 4]]], [1, 2, 3], 5), [1, [1, 2, [1, 2, 3, 5]]]))
    it('inserts a new nested index at end', () =>
        assert.deepStrictEqual(setValueAtKeyPath([1, [2]], [1, -1], 3), [1, [2, 3]]))
    it('overwrites an array', () => assert.deepStrictEqual(setValueAtKeyPath([1, [2]], [1], 3), [1, 3]))
    it('sets a value that requires a new array', () => assert.deepStrictEqual(setValueAtKeyPath([], [0, 0], 1), [[1]]))

    it('sets a nested property (and does not modify input)', () => {
        const input = { a: [{}, { b: [1, 2] }] }
        const origInput = JSON.parse(JSON.stringify(input))
        assert.deepStrictEqual(setValueAtKeyPath(input, ['a', 1, 'b', -1], { c: 3 }), {
            a: [{}, { b: [1, 2, { c: 3 }] }],
        })
        assert.deepStrictEqual(input, origInput)
    })
    it('throws on invalid key type', () => assert.throws(() => setValueAtKeyPath({}, [true as any], {})))
})
