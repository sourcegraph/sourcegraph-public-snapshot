import { initNewExtensionAPI } from '../flatExtentionAPI'
import { SettingsEdit } from '../../client/services/settings'
import { pretendRemote } from '../../util'
import { MainThreadAPI } from '../../contract'

describe('ConfigurationService', () => {
    describe('get()', () => {
        test("throws if initial settings haven't been received", () => {
            const { configuration } = initNewExtensionAPI(pretendRemote({}))
            expect(() => configuration.get()).toThrow('unexpected internal error: settings data is not yet available')
        })
        test('returns the latest settings', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(pretendRemote({}))
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            syncSettingsData({ subjects: [], final: { a: 'c' } })
            expect(configuration.get<{ a: string }>().get('a')).toBe('c')
        })
    })

    describe('changes', () => {
        test('emits as soon as initial settings are received', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(pretendRemote({}))
            let calledTimes = 0
            configuration.subscribe(() => calledTimes++)
            expect(calledTimes).toBe(0)
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            expect(calledTimes).toBe(1)
        })
        test('emits immediately on subscription if initial settings have already been received', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(pretendRemote({}))
            let calledTimes = 0
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            configuration.subscribe(() => calledTimes++)
            expect(calledTimes).toBe(1)
        })

        test('emits when settings are updated', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(pretendRemote({}))
            let calledTimes = 0
            configuration.subscribe(() => calledTimes++)
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            syncSettingsData({ subjects: [], final: { a: 'c' } })
            syncSettingsData({ subjects: [], final: { a: 'd' } })
            expect(calledTimes).toBe(3)
        })

        test('config objects freezes in time??!?!', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(pretendRemote({}))
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            const config = configuration.get<{ a: string }>()
            expect(config.get('a')).toBe('b')
            syncSettingsData({ subjects: [], final: { a: 'c' } })
            expect(config.get('a')).toBe('b') // Shouldn't this be 'c' instead?
        })
    })

    describe('talks to the client api', () => {
        test('talks to the client when an update is requested', async () => {
            const requestedEdits: SettingsEdit[] = []
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(
                pretendRemote<MainThreadAPI>({
                    applySettingsEdit: edit =>
                        Promise.resolve().then(() => {
                            requestedEdits.push(edit)
                        }),
                })
            )
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            const config = configuration.get<{ a: string }>()
            await config.update('a', 'aha!')
            expect(requestedEdits).toEqual<SettingsEdit[]>([{ path: ['a'], value: 'aha!' }])
            expect(config.get('a')).toBe('b') // no optimistic updates
        })
    })
})
