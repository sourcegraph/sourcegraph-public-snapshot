import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'

import { RegistryExtensionFieldsForList } from '../graphql-operations'

import { ConfiguredExtensionCache } from './ExtensionRegistry'
import { configureExtensionRegistry } from './extensions'

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
})
