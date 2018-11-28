import assert from 'assert'
import { from, of, Subscribable } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { ConfiguredExtension } from '../../../extensions/extension'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../../settings/settings'
import { Model } from '../model'
import { ExecutableExtension, ExtensionsService } from './extensionsService'
import { SettingsService } from './settings'

const scheduler = () => new TestScheduler((a, b) => assert.deepStrictEqual(a, b))

class TestExtensionsService extends ExtensionsService {
    constructor(
        mockConfiguredExtensions: ConfiguredExtension[],
        model: Subscribable<Pick<Model, 'visibleTextDocuments'>>,
        settingsService: Pick<SettingsService, 'data'>,
        extensionActivationFilter: (
            enabledExtensions: ConfiguredExtension[],
            model: Pick<Model, 'visibleTextDocuments'>
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
    it('emits an empty set', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    new TestExtensionsService(
                        [],
                        cold<Pick<Model, 'visibleTextDocuments'>>('-a-|', {
                            a: { visibleTextDocuments: [] },
                        }),
                        { data: cold<SettingsCascadeOrError>('-a-|', { a: EMPTY_SETTINGS_CASCADE }) },
                        enabledExtensions => enabledExtensions
                    ).activeExtensions
                )
            ).toBe('-a-|', {
                a: [],
            })
        ))

    it('previously activated extensions remain activated when their activationEvents no longer match', () =>
        scheduler().run(({ cold, expectObservable }) =>
            expectObservable(
                from(
                    new TestExtensionsService(
                        [
                            { id: 'x', manifest: { url: 'u', activationEvents: [] }, rawManifest: null },
                            { id: 'y', manifest: { url: 'u', activationEvents: [] }, rawManifest: null },
                        ],
                        cold<Pick<Model, 'visibleTextDocuments'>>('-a-b-|', {
                            a: { visibleTextDocuments: [{ languageId: 'x', text: '', uri: '' }] },
                            b: { visibleTextDocuments: [{ languageId: 'y', text: '', uri: '' }] },
                        }),
                        {
                            data: cold<SettingsCascadeOrError>('-a-b-|', {
                                a: { final: { extensions: { x: true } }, subjects: [] },
                                b: { final: { extensions: { x: true, y: true } }, subjects: [] },
                            }),
                        },
                        (enabledExtensions, { visibleTextDocuments }) =>
                            enabledExtensions.filter(x =>
                                (visibleTextDocuments || []).some(({ languageId }) => x.id === languageId)
                            )
                    ).activeExtensions
                )
            ).toBe('-a-b-|', {
                a: [{ id: 'x', scriptURL: 'u' }],
                b: [{ id: 'x', scriptURL: 'u' }, { id: 'y', scriptURL: 'u' }],
            } as Record<string, ExecutableExtension[]>)
        ))
})
