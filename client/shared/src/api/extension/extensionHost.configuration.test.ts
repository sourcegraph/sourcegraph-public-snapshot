import { initNewExtensionAPI } from './flatExtensionApi'
import { SettingsEdit } from '../client/services/settings'
import { pretendRemote } from '../util'
import { MainThreadAPI } from '../contract'
import { SettingsCascade } from '../../settings/settings'
import { BehaviorSubject } from 'rxjs'
import { proxySubscribable } from './api/common'
import { proxy } from 'comlink'

const initialSettings = (value: { a: string }): SettingsCascade<{ a: string }> => ({
    subjects: [],
    final: value,
})

describe('ExtensionHost: Configuration', () => {
    describe('get()', () => {
        test('returns the latest settings', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(
                pretendRemote<MainThreadAPI>({
                    getScriptURLForExtension: proxy(() => proxy((scriptURLs: string[]) => Promise.resolve(scriptURLs))),
                    getEnabledExtensions: () => proxySubscribable(new BehaviorSubject([])),
                }),
                {
                    initialSettings: initialSettings({ a: 'a' }),
                    clientApplication: 'sourcegraph',
                }
            )
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            syncSettingsData({ subjects: [], final: { a: 'c' } })
            expect(configuration.get<{ a: string }>().get('a')).toBe('c')
        })
    })

    describe('changes', () => {
        test('emits immediately on subscription', () => {
            const { configuration } = initNewExtensionAPI(
                pretendRemote<MainThreadAPI>({
                    getScriptURLForExtension: proxy(() => proxy((scriptURLs: string[]) => Promise.resolve(scriptURLs))),
                    getEnabledExtensions: () => proxySubscribable(new BehaviorSubject([])),
                }),
                {
                    initialSettings: initialSettings({ a: 'a' }),
                    clientApplication: 'sourcegraph',
                }
            )
            let calledTimes = 0
            configuration.subscribe(() => calledTimes++)
            expect(calledTimes).toBe(1)
        })

        test('emits when settings are updated', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(
                pretendRemote<MainThreadAPI>({
                    getScriptURLForExtension: proxy(() => proxy((scriptURLs: string[]) => Promise.resolve(scriptURLs))),
                    getEnabledExtensions: () => proxySubscribable(new BehaviorSubject([])),
                }),
                {
                    initialSettings: initialSettings({ a: 'a' }),
                    clientApplication: 'sourcegraph',
                }
            )
            let calledTimes = 0
            configuration.subscribe(() => calledTimes++)
            syncSettingsData({ subjects: [], final: { a: 'b' } })
            // one initial and one update
            expect(calledTimes).toBe(2)
        })

        test('config objects freezes in time??!?!', () => {
            const {
                configuration,
                exposedToMain: { syncSettingsData },
            } = initNewExtensionAPI(
                pretendRemote<MainThreadAPI>({
                    getScriptURLForExtension: proxy(() => proxy((scriptURLs: string[]) => Promise.resolve(scriptURLs))),
                    getEnabledExtensions: () => proxySubscribable(new BehaviorSubject([])),
                }),
                {
                    initialSettings: initialSettings({ a: 'b' }),
                    clientApplication: 'sourcegraph',
                }
            )
            const config = configuration.get<{ a: string }>()
            expect(config.get('a')).toBe('b')
            syncSettingsData({ subjects: [], final: { a: 'c' } })
            const newConfigSnapshot = configuration.get<{ a: string }>()
            expect(newConfigSnapshot.get('a')).toBe('c')
        })
    })

    describe('talks to the client api', () => {
        test('talks to the client when an update is requested', async () => {
            const requestedEdits: SettingsEdit[] = []
            const { configuration } = initNewExtensionAPI(
                pretendRemote<MainThreadAPI>({
                    getScriptURLForExtension: proxy(() => proxy((scriptURLs: string[]) => Promise.resolve(scriptURLs))),
                    getEnabledExtensions: () => proxySubscribable(new BehaviorSubject([])),
                    applySettingsEdit: edit =>
                        Promise.resolve().then(() => {
                            requestedEdits.push(edit)
                        }),
                }),
                {
                    initialSettings: initialSettings({ a: 'b' }),
                    clientApplication: 'sourcegraph',
                }
            )
            const config = configuration.get<{ a: string }>()
            await config.update('a', 'aha!')
            expect(requestedEdits).toEqual<SettingsEdit[]>([{ path: ['a'], value: 'aha!' }])
            expect(config.get('a')).toBe('b') // no optimistic updates
        })
    })
})
