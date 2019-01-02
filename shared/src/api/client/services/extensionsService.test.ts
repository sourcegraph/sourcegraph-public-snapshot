import { from, of, Subscribable, throwError } from 'rxjs'
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
        ) => ConfiguredExtension[],
        unpackedExtensionUrl: Subscribable<string>,
        fetchUnpackedExtension: (baseUrl: string) => Subscribable<ConfiguredExtension | null>
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
            extensionActivationFilter,
            fetchUnpackedExtension
        )
        this.configuredExtensions = of(mockConfiguredExtensions)
        this.unpackedExtensionURL = unpackedExtensionUrl as any
    }
}

/* tslint:disable-next-line */
describe.only('activeExtensions', () => {
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
                        enabledExtensions => enabledExtensions,
                        cold('-a-|', { a: '' }),
                        () => of(null)
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
                            ),
                        cold('-a--|', { a: '' }),
                        () => of(null)
                    ).activeExtensions
                )
            ).toBe('-a-b-|', {
                a: [{ id: 'x', manifest, scriptURL: 'u' }],
                b: [{ id: 'x', manifest, scriptURL: 'u' }, { id: 'y', manifest, scriptURL: 'u' }],
            } as Record<string, ExecutableExtension[]>)
        ))

    test('fetches an unpacked extension and adds it to the set of registry extensions', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                from(
                    new TestExtensionsService(
                        [{ id: 'foo', manifest, rawManifest: null }],
                        cold<Pick<Model, 'visibleViewComponents'>>('a-|', {
                            a: { visibleViewComponents: [] },
                        }),
                        {
                            data: cold<SettingsCascadeOrError>('a-|', {
                                a: {
                                    final: {
                                        extensions: {
                                            foo: true,
                                        },
                                    },
                                    subjects: [],
                                },
                            }),
                        },
                        enabledExtensions => enabledExtensions,
                        cold('a-|', { a: 'bar' }),
                        baseUrl =>
                            of({
                                id: baseUrl,
                                manifest: {
                                    url: 'bar.js',
                                    activationEvents: [],
                                },
                                rawManifest: null,
                            })
                    ).activeExtensions
                )
            ).toBe('a-|', {
                a: [
                    { id: 'foo', manifest, scriptURL: 'u' },
                    { id: 'bar', manifest: { url: 'bar.js', activationEvents: [] }, scriptURL: 'bar.js' },
                ],
            })
        })
    })

    test('still returns registry extensions even if fetching an unpacked extension fails', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                from(
                    new TestExtensionsService(
                        [{ id: 'foo', manifest, rawManifest: null }],
                        cold<Pick<Model, 'visibleViewComponents'>>('a-|', {
                            a: { visibleViewComponents: [] },
                        }),
                        {
                            data: cold<SettingsCascadeOrError>('a-|', {
                                a: {
                                    final: {
                                        extensions: {
                                            foo: true,
                                        },
                                    },
                                    subjects: [],
                                },
                            }),
                        },
                        enabledExtensions => enabledExtensions,
                        cold('a-|', { a: 'bar' }),
                        () => throwError('baz')
                    ).activeExtensions
                )
            ).toBe('a-|', {
                a: [{ id: 'foo', manifest, scriptURL: 'u' }],
            })
        })
    })
})
