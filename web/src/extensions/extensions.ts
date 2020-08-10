import { RegistryExtensionFieldsForList } from '../graphql-operations'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import { ConfiguredRegistryExtension, toConfiguredRegistryExtension } from '../../../shared/src/extensions/extension'
import { validCategories } from './extension/extension'
import { isErrorLike } from '../../../shared/src/util/errors'

export interface CategorizedExtensionRegistry {
    /** Maps categories to ids of extensions  */
    categories: Record<ExtensionCategory, string[]>

    /** All extensions returned by the query indexed by id */
    extensions: { [id: string]: ConfiguredRegistryExtension<RegistryExtensionFieldsForList> }
}

const NO_VALID_CATEGORIES: 'Other'[] = ['Other']
/**
 * Normalizes
 *
 *
 */
export function categorizeExtensionRegistry(
    nodes: RegistryExtensionFieldsForList[],
    configuredExtensionsCache: Map<string, ConfiguredRegistryExtension<RegistryExtensionFieldsForList>>
): CategorizedExtensionRegistry {
    const categoriesById = EXTENSION_CATEGORIES.reduce((categories, category) => {
        categories[category] = []
        return categories
        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    }, {} as Record<ExtensionCategory, string[]>)

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
