import { from, of, Subscribable } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { ConfiguredExtension } from '../../../extensions/extension'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../../settings/settings'
import { Model } from '../model'
import { ExecutableExtension, ExtensionsService } from './extensionsService'
import { SettingsService } from './settings'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

class TestExtensionsService extends ExtensionsService {
    constructor(
        mockConfiguredExtensions: ConfiguredExtension[],
        model: Subscribable<Pick<Model, 'visibleViewComponents'>>,
        settingsService: Pick<SettingsService, 'data'>,
        extensionActivationFilter: (
            enabledExtensions: ConfiguredExtension[],
            model: Pick<Model, 'visibleViewComponents'>
        ) => ConfiguredExtension[]
    ) {
        super(
            {
                queryGraphQL: () => {
                    throw new Error('not implemented')
                },
                getScriptURLForExtension: scriptURL => scriptURL,
            },
            model,
            settingsService,
            extensionActivationFilter
        )
        this.configuredExtensions = of(mockConfiguredExtensions)
    }
}

describe('activeExtensions', () => {
    test('emits an empty set', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    new TestExtensionsService(
                        [],
                        cold<Pick<Model, 'visibleViewComponents'>>('-a-|', {
                            a: { visibleViewComponents: [] },
                        }),
                        { data: cold<SettingsCascadeOrError>('-a-|', { a: EMPTY_SETTINGS_CASCADE }) },
                        enabledExtensions => enabledExtensions
                    ).activeExtensions
                )
            ).toBe('-a-|', {
                a: [],
            })
        ))

    const manifest = { url: 'u', activationEvents: [] }
    test('previously activated extensions remain activated when their activationEvents no longer match', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    new TestExtensionsService(
                        [{ id: 'x', manifest, rawManifest: null }, { id: 'y', manifest, rawManifest: null }],
                        cold<Pick<Model, 'visibleViewComponents'>>('-a-b-|', {
                            a: {
                                visibleViewComponents: [
                                    {
                                        type: 'textEditor',
                                        item: { languageId: 'x', text: '', uri: '' },
                                        selections: [],
                                        isActive: true,
                                    },
                                ],
                            },
                            b: {
                                visibleViewComponents: [
                                    {
                                        type: 'textEditor',
                                        item: { languageId: 'y', text: '', uri: '' },
                                        selections: [],
                                        isActive: true,
                                    },
                                ],
                            },
                        }),
                        {
                            data: cold<SettingsCascadeOrError>('-a-b-|', {
                                a: { final: { extensions: { x: true } }, subjects: [] },
                                b: { final: { extensions: { x: true, y: true } }, subjects: [] },
                            }),
                        },
                        (enabledExtensions, { visibleViewComponents }) =>
                            enabledExtensions.filter(x =>
                                (visibleViewComponents || []).some(({ item: { languageId } }) => x.id === languageId)
                            )
                    ).activeExtensions
                )
            ).toBe('-a-b-|', {
                a: [{ id: 'x', manifest, scriptURL: 'u' }],
                b: [{ id: 'x', manifest, scriptURL: 'u' }, { id: 'y', manifest, scriptURL: 'u' }],
            } as Record<string, ExecutableExtension[]>)
        ))
})
