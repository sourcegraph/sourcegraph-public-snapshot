import { assert } from 'chai'
import { ConfigurationUpdateParams } from 'sourcegraph/module/protocol'
import { convertUpdateConfigurationCommandArgs } from './clientCommands'

describe('convertUpdateConfigurationCommandArgs', () => {
    it('converts with a non-JSON-encoded arg', () =>
        assert.deepEqual(convertUpdateConfigurationCommandArgs([['a', 1], { x: 2 }]), {
            path: ['a', 1],
            value: { x: 2 },
        } as ConfigurationUpdateParams))

    it('converts with a JSON-encoded arg', () =>
        assert.deepEqual(convertUpdateConfigurationCommandArgs([['a', 1], '"x"', null, 'json']), {
            path: ['a', 1],
            value: 'x',
        } as ConfigurationUpdateParams))

    it('throws if the arg is invalid', () => assert.throws(() => convertUpdateConfigurationCommandArgs([] as any)))
})
