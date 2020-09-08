import { RegistryExtensionFieldsForList } from '../graphql-operations'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import {
    ConfiguredRegistryExtension,
    toConfiguredRegistryExtension,
    isExtensionEnabled,
} from '../../../shared/src/extensions/extension'
import { validCategories } from './extension/extension'
import { isErrorLike, ErrorLike } from '../../../shared/src/util/errors'
import { ExtensionsEnablement } from './ExtensionRegistry'
import { Settings } from '../../../shared/src/settings/settings'

export interface CategorizedExtensionRegistry {
    /** Maps categories to ids of extensions  */
    categories: Record<ExtensionCategory, string[]>

    /** All extensions returned by the query indexed by id */
    extensions: { [id: string]: ConfiguredRegistryExtension<RegistryExtensionFieldsForList> }
}

const NO_VALID_CATEGORIES: 'Other'[] = ['Other']

/**
 * Groups registry extensions by category.
 *
 * `categories`: Object mapping category name to array of extension ids in that category
 * `extensions`: Object mapping extension id to the configured extension with that id
 *
 * `categorizeExtensionRegistry` is passed a cache of configured extensions to avoid
 * parsing manifests multiple times during the lifecycle of the extension registry.
 *
 */
export function categorizeExtensionRegistry(
    nodes: RegistryExtensionFieldsForList[],
    configuredExtensionsCache: Map<string, ConfiguredRegistryExtension<RegistryExtensionFieldsForList>>
): CategorizedExtensionRegistry {
    const categoriesById: Record<string, string[]> = {}

    for (const category of EXTENSION_CATEGORIES) {
        categoriesById[category] = []
    }

    const categorizedExtensionRegistry: CategorizedExtensionRegistry = {
        categories: categoriesById,
        extensions: {},
    }

    for (const node of nodes) {
        // cache parsed extension manifests
        let configuredRegistryExtension = configuredExtensionsCache.get(node.id)
        if (!configuredRegistryExtension) {
            configuredRegistryExtension = toConfiguredRegistryExtension(node)
            configuredExtensionsCache.set(node.id, configuredRegistryExtension)
        }

        let categories: ExtensionCategory[]
        if (!isErrorLike(configuredRegistryExtension.manifest) && configuredRegistryExtension.manifest?.categories) {
            categories = validCategories(configuredRegistryExtension.manifest.categories) || NO_VALID_CATEGORIES
        } else {
            categories = NO_VALID_CATEGORIES
        }
        const primaryCategory = categories[0]
        categorizedExtensionRegistry.categories[primaryCategory].push(configuredRegistryExtension.id)
        categorizedExtensionRegistry.extensions[configuredRegistryExtension.id] = configuredRegistryExtension
    }

    return categorizedExtensionRegistry
}

/**
 * Filters categorized registry extensions by enablement (enabled | disabled | all)
 */
export function applyExtensionsEnablement(
    categories: CategorizedExtensionRegistry['categories'],
    filteredCategoryIDs: ExtensionCategory[],
    enablement: ExtensionsEnablement,
    settings: Settings | ErrorLike | null
): CategorizedExtensionRegistry['categories'] {
    if (enablement === 'all') {
        return categories
    }

    const enabled = enablement === 'enabled'

    const filteredCategories: Record<string, string[]> = {}

    for (const category of filteredCategoryIDs) {
        filteredCategories[category] = categories[category].filter(
            extensionID => enabled === isExtensionEnabled(settings, extensionID)
        )
    }

    return filteredCategories
}
