import { SettingsEdit } from '../api/client/services/settings'
import { convertUpdateConfigurationCommandArgs } from './commands'

describe('convertUpdateConfigurationCommandArgs', () => {
    test('converts with a non-JSON-encoded arg', () =>
        expect(convertUpdateConfigurationCommandArgs([['a', 1], { x: 2 }])).toEqual({
            path: ['a', 1],
            value: { x: 2 },
        } as SettingsEdit))

    test('converts with a JSON-encoded arg', () =>
        expect(convertUpdateConfigurationCommandArgs([['a', 1], '"x"', null, 'json'])).toEqual({
            path: ['a', 1],
            value: 'x',
        } as SettingsEdit))

    test('throws if the arg is invalid', () => expect(() => convertUpdateConfigurationCommandArgs([] as any)).toThrow())
})
