import { initMainThreadAPI } from './mainthreadAPI'
import { PlatformContext } from '../../platform/context'
import { of, Subject } from 'rxjs'
import { pretendRemote } from '../util'
import { SettingsEdit } from './services/settings'
import { SettingsCascade } from '../../settings/settings'
import { FlatExtHostAPI } from '../contract'

describe('configuration', () => {
    test('changeConfiguration goes to platform with the last settings subject', async () => {
        let calledWith: Parameters<PlatformContext['updateSettings']> | undefined
        const updateSettings: PlatformContext['updateSettings'] = (...args) => {
            calledWith = args
            return Promise.resolve()
        }
        const platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'> = {
            settings: of({
                subjects: [
                    {
                        settings: null,
                        lastID: null,
                        subject: { id: 'id1', __typename: 'DefaultSettings', viewerCanAdminister: true },
                    },
                    {
                        settings: null,
                        lastID: null,

                        subject: { id: 'id2', __typename: 'DefaultSettings', viewerCanAdminister: true },
                    },
                ],
                final: { a: 'value' },
            }),
            updateSettings,
        }

        const [api] = initMainThreadAPI(pretendRemote({}), platformContext)

        const edit: SettingsEdit = { path: ['a'], value: 'newVal' }
        await api.applySettingsEdit(edit)

        expect(calledWith).toEqual<Parameters<PlatformContext['updateSettings']>>(['id2', edit])
    })

    test('changes of settings from platform propagated to the ext host', () => {
        const values: SettingsCascade<{ a: string }>[] = [
            {
                subjects: [], // this is valid actually even though it shouldn't
                final: { a: 'one' },
            },
            {
                subjects: null as any, // invalid and should be ignored
                final: { a: 'invalid two' },
            },
            {
                subjects: [], // one more valid case
                final: { a: 'three' },
            },
        ]

        const platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'> = {
            settings: of(...values),
            updateSettings: () => Promise.resolve(),
        }

        const passedToExtHost: SettingsCascade<object>[] = []
        initMainThreadAPI(
            pretendRemote<FlatExtHostAPI>({
                syncSettingsData: data => {
                    passedToExtHost.push(data)
                },
            }),
            platformContext
        )

        expect(passedToExtHost).toEqual<SettingsCascade<{ a: string }>[]>([values[0], values[2]])
    })

    test('changes of settings are not passed to ext host after unsub', () => {
        const values = new Subject<SettingsCascade<{ a: string }>>()
        const platformContext: Pick<PlatformContext, 'updateSettings' | 'settings'> = {
            settings: values.asObservable(),
            updateSettings: () => Promise.resolve(),
        }

        const passedToExtHost: SettingsCascade<object>[] = []
        const [, subscription] = initMainThreadAPI(
            pretendRemote<FlatExtHostAPI>({
                syncSettingsData: data => {
                    passedToExtHost.push(data)
                },
            }),
            platformContext
        )

        const one = {
            subjects: [],
            final: { a: 'one' },
        }

        const two = {
            subjects: [],
            final: { a: 'two' },
        }

        values.next(one)
        expect(passedToExtHost).toEqual([one])

        subscription.unsubscribe()
        values.next(two)
        expect(passedToExtHost).toEqual([one]) // nothing happened after unsub
    })
})
