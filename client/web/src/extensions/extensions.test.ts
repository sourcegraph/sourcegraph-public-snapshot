import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'
import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { RegistryExtensionFieldsForList } from '../graphql-operations'

import { ConfiguredExtensionCache } from './ExtensionRegistry'
import {
    applyEnablementFilter,
    applyWIPFilter,
    ConfiguredRegistryExtensions,
    configureExtensionRegistry,
} from './extensions'

describe('extension registry helpers', () => {
    describe('configureExtensionRegistry', () => {
        test('should group extensions by first listed category', () => {
            const configuredExtensionCache: ConfiguredExtensionCache = new Map<
                string,
                ConfiguredRegistryExtension<RegistryExtensionFieldsForList>
            >()

            const nodes: RegistryExtensionFieldsForList[] = [
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                { id: 'sourcegraph/snyk' } as RegistryExtensionFieldsForList,
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                { id: 'sourcegraph/eslint' } as RegistryExtensionFieldsForList,
            ]

            configuredExtensionCache.set('sourcegraph/snyk', {
                id: 'sourcegraph/snyk',
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                manifest: {
                    categories: ['External services', 'Code analysis'],
                } as ExtensionManifest,
            })

            configuredExtensionCache.set('sourcegraph/eslint', {
                id: 'sourcegraph/eslint',
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                manifest: {
                    categories: ['Code analysis', 'Linters'],
                } as ExtensionManifest,
            })

            const { extensionIDsByCategory } = configureExtensionRegistry(nodes, configuredExtensionCache)
            expect(extensionIDsByCategory['External services'].primaryExtensionIDs).toStrictEqual(['sourcegraph/snyk'])
            expect(extensionIDsByCategory['Code analysis'].primaryExtensionIDs).toStrictEqual(['sourcegraph/eslint'])
        })
    })

    describe('applyEnablementFilter', () => {
        const extensionIDs = ['sourcegraph/codecov', 'sourcegraph/snyk', 'sourcegraph/eslint']
        const settings: Settings = {
            extensions: {
                'sourcegraph/codecov': true,
                'sourcegraph/eslint': false,
            },
        }

        test('returns all extensionIDs when filter is "all"', () => {
            expect(applyEnablementFilter(extensionIDs, 'all', settings)).toStrictEqual([
                'sourcegraph/codecov',
                'sourcegraph/snyk',
                'sourcegraph/eslint',
            ])
        })
        test('returns only enabled extensionIDs when filter is "enabled"', () => {
            expect(applyEnablementFilter(extensionIDs, 'enabled', settings)).toStrictEqual(['sourcegraph/codecov'])
        })
        test('returns only disabled extensionIDs when filter is "disabled"', () => {
            expect(applyEnablementFilter(extensionIDs, 'disabled', settings)).toStrictEqual([
                'sourcegraph/snyk',
                'sourcegraph/eslint',
            ])
        })
    })

    describe('appyWIPFilter', () => {
        const extensionIDs = ['sourcegraph/codecov', 'sourcegraph/snyk', 'sourcegraph/eslint']
        const extensions: ConfiguredRegistryExtensions = {
            'sourcegraph/codecov': {
                id: 'sourcegraph/codecov',
                manifest: {} as ExtensionManifest,
            },
            'sourcegraph/snyk': {
                id: 'sourcegraph/snyk',
                manifest: {
                    wip: true,
                } as ExtensionManifest,
            },
            'sourcegraph/eslint': {
                id: 'sourcegraph/eslint',
                manifest: {
                    wip: false,
                } as ExtensionManifest,
            },
        }

        test('returns all extensionIDs when filter is `true`', () => {
            expect(applyWIPFilter(extensionIDs, true, extensions)).toStrictEqual([
                'sourcegraph/codecov',
                'sourcegraph/snyk',
                'sourcegraph/eslint',
            ])
        })

        test('returns only non-wip extensions when filter is `false`', () => {
            expect(applyWIPFilter(extensionIDs, false, extensions)).toStrictEqual([
                'sourcegraph/codecov',
                'sourcegraph/eslint',
            ])
        })
    })
})
