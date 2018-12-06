import * as assert from 'assert'
import { BehaviorSubject, of } from 'rxjs'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../settings/settings'
import { SettingsEdit } from '../client/services/settings'
import { assertToJSON } from '../extension/types/common.test'
import { integrationTestContext } from './helpers.test'

describe('Configuration (integration)', () => {
    it('is synchronously available', async () => {
        const { extensionHost } = await integrationTestContext({ settings: of(EMPTY_SETTINGS_CASCADE) })
        assert.doesNotThrow(() => extensionHost.configuration.subscribe(() => void 0))
        assert.doesNotThrow(() => extensionHost.configuration.get())
    })

    describe('Configuration#get', () => {
        it('gets configuration', async () => {
            const { extensionHost } = await integrationTestContext({ settings: of({ final: { a: 1 }, subjects: [] }) })
            assertToJSON(extensionHost.configuration.get(), { a: 1 })
            assert.deepStrictEqual(extensionHost.configuration.get().value, { a: 1 })
        })
    })

    describe('Configuration#update', () => {
        it('updates configuration', async () => {
            const calls: (SettingsEdit | string)[] = []
            const { extensionHost } = await integrationTestContext({
                settings: of({ final: { a: 1 }, subjects: [{ subject: {} as any, lastID: null, settings: null }] }),
                updateSettings: async (_subject, edit) => {
                    calls.push(edit)
                },
            })

            await extensionHost.configuration.get().update('a', 2)
            await extensionHost.internal.sync()
            assert.deepStrictEqual(calls, [{ path: ['a'], value: 2 }] as SettingsEdit[])
            calls.length = 0 // clear

            await extensionHost.configuration.get().update('a', 3)
            await extensionHost.internal.sync()
            assert.deepStrictEqual(calls, [{ path: ['a'], value: 3 }] as SettingsEdit[])
        })
    })

    describe('configuration.subscribe', () => {
        it('subscribes to changes', async () => {
            const mockSettings = new BehaviorSubject<SettingsCascadeOrError>(EMPTY_SETTINGS_CASCADE)
            const { extensionHost } = await integrationTestContext({ settings: mockSettings })

            let calls = 0
            extensionHost.configuration.subscribe(() => calls++)
            assert.strictEqual(calls, 1) // called initially

            mockSettings.next(EMPTY_SETTINGS_CASCADE)
            await extensionHost.internal.sync()
            assert.strictEqual(calls, 2)
        })
    })
})
