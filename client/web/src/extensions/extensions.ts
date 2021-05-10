import {
    ConfiguredRegistryExtension,
    toConfiguredRegistryExtension,
} from '@sourcegraph/shared/src/extensions/extension'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'
import { createRecord } from '@sourcegraph/shared/src/util/createRecord'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { RegistryExtensionFieldsForList } from '../graphql-operations'

import { validCategories } from './extension/extension'
import { ConfiguredExtensionCache } from './ExtensionRegistry'

export interface ConfiguredRegistryExtensions {
    [id: string]: Pick<ConfiguredRegistryExtension<RegistryExtensionFieldsForList>, 'manifest' | 'id'>
}

export interface ConfiguredExtensionRegistry {
    /** Maps categories to ids of extensions  */
    extensionIDsByCategory: Record<
        ExtensionCategory,
        {
            /** IDs of all extensions for which this is the primary category */
            primaryExtensionIDs: string[]
            /** IDs of all extensions that fall into this category */
            allExtensionIDs: string[]
        }
    >

    /** All extensions returned by the query indexed by id */
    extensions: ConfiguredRegistryExtensions
}

const NO_VALID_CATEGORIES: 'Other'[] = ['Other']

/**
 * Configures extensions for the registry.
 *
 * `extensionIDsByCategory`: Object mapping category name to array of extension ids in that category
 * `extensions`: Object mapping extension id to the configured extension with that id
 *
 * `configureExtensionRegistry` is passed a cache of configured extensions to avoid
 * parsing manifests multiple times during the lifecycle of the extension registry.
 */
export function configureExtensionRegistry(
    nodes: RegistryExtensionFieldsForList[],
    configuredExtensionCache: ConfiguredExtensionCache
): ConfiguredExtensionRegistry {
    const extensions: ConfiguredRegistryExtensions = {}

    const extensionIDsByCategory: ConfiguredExtensionRegistry['extensionIDsByCategory'] = createRecord(
        EXTENSION_CATEGORIES,
        () => ({ primaryExtensionIDs: [], allExtensionIDs: [] })
    )

    for (const node of nodes) {
        // cache parsed extension manifests
        let configuredRegistryExtension = configuredExtensionCache.get(node.id)
        if (!configuredRegistryExtension) {
            configuredRegistryExtension = toConfiguredRegistryExtension(node)
            configuredExtensionCache.set(node.id, configuredRegistryExtension)
        }

        let categories: ExtensionCategory[]
        if (!isErrorLike(configuredRegistryExtension.manifest) && configuredRegistryExtension.manifest?.categories) {
            categories = validCategories(configuredRegistryExtension.manifest.categories) || NO_VALID_CATEGORIES
        } else {
            categories = NO_VALID_CATEGORIES
        }
        // TODO: Add `primaryCategory` to extension schema
        // Primary category is either specified or inferred by array position
        const primaryCategory = categories[0]

        extensionIDsByCategory[primaryCategory].primaryExtensionIDs.push(configuredRegistryExtension.id)
        for (const category of categories) {
            extensionIDsByCategory[category].allExtensionIDs.push(configuredRegistryExtension.id)
        }

        extensions[configuredRegistryExtension.id] = configuredRegistryExtension
    }

    return { extensions, extensionIDsByCategory }
}
