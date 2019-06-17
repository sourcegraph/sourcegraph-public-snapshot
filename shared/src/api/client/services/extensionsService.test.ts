import { from, of, Subscribable, throwError } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { ConfiguredExtension } from '../../../extensions/extension'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../../settings/settings'
import { CodeEditorWithPartialModel, EditorService } from './editorService'
import { createTestEditorService } from './editorService.test'
import { ExecutableExtension, ExtensionsService } from './extensionsService'
import { SettingsService } from './settings'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

class TestExtensionsService extends ExtensionsService {
    constructor(
        mockConfiguredExtensions: ConfiguredExtension[],
        editorService: Pick<EditorService, 'editorsAndModels'>,
        settingsService: Pick<SettingsService, 'data'>,
        extensionActivationFilter: (
            enabledExtensions: ConfiguredExtension[],
            editors: readonly CodeEditorWithPartialModel[]
        ) => ConfiguredExtension[],
        sideloadedExtensionURL: Subscribable<string | null>,
        fetchSideloadedExtension: (baseUrl: string) => Subscribable<ConfiguredExtension | null>
    ) {
        super(
            {
                requestGraphQL: () => {
                    throw new Error('not implemented')
                },
                getScriptURLForExtension: scriptURL => scriptURL,
                sideloadedExtensionURL,
            },
            editorService,
            settingsService,
            extensionActivationFilter,
            fetchSideloadedExtension
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
                        createTestEditorService(
                            cold<readonly CodeEditorWithPartialModel[]>('-a-|', {
                                a: [],
                            })
                        ),
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
                        createTestEditorService(
                            cold<readonly CodeEditorWithPartialModel[]>('-a-b-|', {
                                a: [
                                    {
                                        type: 'CodeEditor',
                                        editorId: 'editor#0',
                                        resource: 'u',
                                        model: { languageId: 'x' },
                                        selections: [],
                                        isActive: true,
                                    },
                                ],
                                b: [
                                    {
                                        type: 'CodeEditor',
                                        editorId: 'editor#1',
                                        resource: 'u2',
                                        model: { languageId: 'y' },
                                        selections: [],
                                        isActive: true,
                                    },
                                ],
                            })
                        ),
                        {
                            data: cold<SettingsCascadeOrError>('-a-b-|', {
                                a: { final: { extensions: { x: true } }, subjects: [] },
                                b: { final: { extensions: { x: true, y: true } }, subjects: [] },
                            }),
                        },
                        (enabledExtensions, editors) =>
                            enabledExtensions.filter(x =>
                                editors.some(({ model: { languageId } }) => x.id === languageId)
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

    test('fetches a sideloaded extension and adds it to the set of registry extensions', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                from(
                    new TestExtensionsService(
                        [{ id: 'foo', manifest, rawManifest: null }],
                        createTestEditorService(
                            cold<readonly CodeEditorWithPartialModel[]>('a-|', {
                                a: [],
                            })
                        ),
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

    test('still returns registry extensions even if fetching a sideloaded extension fails', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                from(
                    new TestExtensionsService(
                        [{ id: 'foo', manifest, rawManifest: null }],
                        createTestEditorService(
                            cold<readonly CodeEditorWithPartialModel[]>('a-|', {
                                a: [],
                            })
                        ),
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
                        () => throwError(new Error('baz'))
                    ).activeExtensions
                )
            ).toBe('a-|', {
                a: [{ id: 'foo', manifest, scriptURL: 'u' }],
            })
        })
    })
})
