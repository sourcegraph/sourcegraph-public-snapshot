import { ErrorLike, isErrorLike, isDefined } from '@sourcegraph/common'
import {
    ConfiguredRegistryExtension,
    isExtensionEnabled,
    toConfiguredRegistryExtension,
} from '@sourcegraph/shared/src/extensions/extension'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'
import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { RegistryExtensionFieldsForList } from '../graphql-operations'

import { validCategories } from './extension/extension'
import { ConfiguredExtensionCache, ExtensionsEnablement } from './ExtensionRegistry'
import { createRecord } from './utils/createRecord'

export type MinimalConfiguredRegistryExtension = Pick<
    ConfiguredRegistryExtension<RegistryExtensionFieldsForList>,
    'manifest' | 'id'
>

export interface ConfiguredRegistryExtensions {
    [id: string]: MinimalConfiguredRegistryExtension
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

/**
 * Configures featured extensions to be displayed on the extension registry.
 *
 * Share configured extension cache with `configureExtensionRegistry`
 * since featured extensions are likely to be displayed twice on the page.
 */
export function configureFeaturedExtensions(
    featuredExtensions: RegistryExtensionFieldsForList[],
    configuredExtensionCache: ConfiguredExtensionCache
): MinimalConfiguredRegistryExtension[] {
    const extensions: MinimalConfiguredRegistryExtension[] = []

    for (const featuredExtension of featuredExtensions) {
        let configuredRegistryExtension = configuredExtensionCache.get(featuredExtension.id)
        if (!configuredRegistryExtension) {
            configuredRegistryExtension = toConfiguredRegistryExtension(featuredExtension)
            configuredExtensionCache.set(featuredExtension.id, configuredRegistryExtension)
        }
        extensions.push(configuredRegistryExtension)
    }

    return extensions
}

/**
 * Removes extensions that do not satify the enablement filter.
 *
 * For example, if the user wants to see only enabled extensions, remove disabled extensions.
 */
export function applyEnablementFilter(
    extensionIDs: string[],
    enablementFilter: ExtensionsEnablement,
    settings: Settings | ErrorLike | null
): string[] {
    if (enablementFilter === 'all') {
        return extensionIDs
    }

    return extensionIDs.filter(extensionID => {
        const showEnabled = enablementFilter === 'enabled'
        const isEnabled = isExtensionEnabled(settings, extensionID)

        return showEnabled === isEnabled
    })
}

/**
 * Removes extensions that do not satisfy the WIP/experimental filter.
 *
 * For example, if the user does not want to see experimental extensions,
 * remove all extensions where WIP === true.
 */
export function applyWIPFilter(
    extensionIDs: string[],
    wipFilter: boolean,
    extensions: ConfiguredRegistryExtensions
): string[] {
    if (wipFilter === true) {
        return extensionIDs
    }

    return extensionIDs.filter(extensionID => {
        const extension = extensions[extensionID]
        if (!extension) {
            return false // Shouldn't be reached
        }

        if (extension.manifest && !isErrorLike(extension.manifest) && isDefined(extension.manifest.wip)) {
            return !extension.manifest.wip // Don't include WIP extensions
        }

        // Don't filter it out if we don't have enough information to determine
        // that the extension is WIP.
        return true
    })
}
