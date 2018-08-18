import assert from 'assert'
import { NotificationHandler } from '../../jsonrpc2/handlers'
import {
    ConfigurationCascade,
    ConfigurationUpdateParams,
    ConfigurationUpdateRequest,
    DidChangeConfigurationParams,
    Settings,
} from '../../protocol'
import { Connection } from '../server'
import { RemoteConfiguration, setValueAtKeyPath } from './configuration'

const EMPTY_MOCK_CONNECTION: Connection = {
    onDidChangeConfiguration: () => void 0,
    sendRequest: () => Promise.resolve(void 0),
} as any

const FIXTURE_CONFIGURATION_CASCADE: ConfigurationCascade<Settings> = { merged: { a: 1 } }

describe('RemoteConfigurationImpl', () => {
    const create = (
        connection: Connection = EMPTY_MOCK_CONNECTION,
        configurationCascade = FIXTURE_CONFIGURATION_CASCADE
    ): RemoteConfiguration<Settings> => {
        const remote = new RemoteConfiguration<Settings>()
        remote.attach(connection)
        remote.initialize({
            root: null,
            capabilities: {},
            configurationCascade,
            workspaceFolders: null,
        })
        return remote
    }

    it('initially reports an empty config', () => {
        const remote = new RemoteConfiguration<Settings>()
        assert.deepStrictEqual(remote.configuration.value, {} as Settings)
    })

    it('records the configuration value from initialize', () => {
        const remote = create()
        assert.deepStrictEqual(remote.configuration.value, { a: 1 } as Settings)
    })

    it('records the configuration value from client update notifications', () => {
        let onDidChangeConfiguration: NotificationHandler<DidChangeConfigurationParams> | undefined
        const remote = create({
            ...EMPTY_MOCK_CONNECTION,
            onDidChangeConfiguration: h => (onDidChangeConfiguration = h),
        })
        assert.ok(onDidChangeConfiguration)
        onDidChangeConfiguration!({ configurationCascade: { merged: { b: 2 } } })
        assert.deepStrictEqual(remote.configuration.value, { b: 2 } as Settings)
    })

    it('updateConfiguration edit is sent as request and reflected immediately', async () => {
        const remote = create({
            ...EMPTY_MOCK_CONNECTION,
            sendRequest: async (type: any, params: any) => {
                assert.strictEqual(type, ConfigurationUpdateRequest.type)
                assert.deepStrictEqual(params, { path: ['b'], value: 2 } as ConfigurationUpdateParams)
            },
        })
        const updated = remote.updateConfiguration(['b'], 2)
        assert.deepStrictEqual(remote.configuration.value, { a: 1, b: 2 } as Settings)
        await updated
        assert.deepStrictEqual(remote.configuration.value, { a: 1, b: 2 } as Settings)
    })

    it('handles interleaved updateConfiguration and didChangeConfiguration (authoritative)', async () => {
        let onDidChangeConfiguration: NotificationHandler<DidChangeConfigurationParams> | undefined
        const remote = create({
            ...EMPTY_MOCK_CONNECTION,
            sendRequest: async (type: any, params: any) => {
                assert.strictEqual(type, ConfigurationUpdateRequest.type)
                assert.deepStrictEqual(params, { path: ['b'], value: 2 } as ConfigurationUpdateParams)
            },
            onDidChangeConfiguration: h => (onDidChangeConfiguration = h),
        })
        const updated = remote.updateConfiguration(['b'], 2)
        assert.deepStrictEqual(remote.configuration.value, { a: 1, b: 2 } as Settings)
        onDidChangeConfiguration!({ configurationCascade: { merged: { c: 3 } } })
        await updated
        assert.deepStrictEqual(remote.configuration.value, { c: 3 } as Settings)
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
