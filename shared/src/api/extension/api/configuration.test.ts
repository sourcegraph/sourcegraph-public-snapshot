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
                exposedToMain: { updateConfigurationData },
            } = initNewExtensionAPI(pretendRemote({}))
            updateConfigurationData({ subjects: [], final: { a: 'b' } })
            updateConfigurationData({ subjects: [], final: { a: 'c' } })
            expect(configuration.get<{ a: string }>().get('a')).toBe('c')
        })
    })

    describe('changes', () => {
        test('emits as soon as initial settings are received', () => {
            const {
                configuration,
                exposedToMain: { updateConfigurationData },
            } = initNewExtensionAPI(pretendRemote({}))
            let calledTimes = 0
            configuration.subscribe(() => calledTimes++)
            expect(calledTimes).toBe(0)
            updateConfigurationData({ subjects: [], final: { a: 'b' } })
            expect(calledTimes).toBe(1)
        })
        test('emits immediately on subscription if initial settings have already been received', () => {
            const {
                configuration,
                exposedToMain: { updateConfigurationData },
            } = initNewExtensionAPI(pretendRemote({}))
            let calledTimes = 0
            updateConfigurationData({ subjects: [], final: { a: 'b' } })
            configuration.subscribe(() => calledTimes++)
            expect(calledTimes).toBe(1)
        })

        test('emits when settings are updated', () => {
            const {
                configuration,
                exposedToMain: { updateConfigurationData },
            } = initNewExtensionAPI(pretendRemote({}))
            let calledTimes = 0
            configuration.subscribe(() => calledTimes++)
            updateConfigurationData({ subjects: [], final: { a: 'b' } })
            updateConfigurationData({ subjects: [], final: { a: 'c' } })
            updateConfigurationData({ subjects: [], final: { a: 'd' } })
            expect(calledTimes).toBe(3)
        })

        test('config objects freezes in time??!?!', () => {
            const {
                configuration,
                exposedToMain: { updateConfigurationData },
            } = initNewExtensionAPI(pretendRemote({}))
            updateConfigurationData({ subjects: [], final: { a: 'b' } })
            const config = configuration.get<{ a: string }>()
            expect(config.get('a')).toBe('b')
            updateConfigurationData({ subjects: [], final: { a: 'c' } })
            expect(config.get('a')).toBe('b') // Shouldn't this be 'c' instead?
        })
    })

    describe('talks to the client api', () => {
        test('talks to the client when an update is requested', async () => {
            const requestedEdits: SettingsEdit[] = []
            const {
                configuration,
                exposedToMain: { updateConfigurationData },
            } = initNewExtensionAPI(
                pretendRemote<MainThreadAPI>({
                    changeConfiguration: edit =>
                        Promise.resolve().then(() => {
                            requestedEdits.push(edit)
                        }),
                })
            )
            updateConfigurationData({ subjects: [], final: { a: 'b' } })
            const config = configuration.get<{ a: string }>()
            await config.update('a', 'aha!')
            expect(requestedEdits).toEqual<SettingsEdit[]>([{ path: ['a'], value: 'aha!' }])
            expect(config.get('a')).toBe('b') // no optimistic updates
        })
    })
})
