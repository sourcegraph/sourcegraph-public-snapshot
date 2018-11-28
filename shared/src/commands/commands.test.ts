import assert from 'assert'
import { SettingsEdit } from '../api/client/services/settings'
import { convertUpdateConfigurationCommandArgs } from './commands'

describe('convertUpdateConfigurationCommandArgs', () => {
    it('converts with a non-JSON-encoded arg', () =>
        assert.deepStrictEqual(convertUpdateConfigurationCommandArgs([['a', 1], { x: 2 }]), {
            path: ['a', 1],
            value: { x: 2 },
        } as SettingsEdit))

    it('converts with a JSON-encoded arg', () =>
        assert.deepStrictEqual(convertUpdateConfigurationCommandArgs([['a', 1], '"x"', null, 'json']), {
            path: ['a', 1],
            value: 'x',
        } as SettingsEdit))

    it('throws if the arg is invalid', () => assert.throws(() => convertUpdateConfigurationCommandArgs([] as any)))
})
