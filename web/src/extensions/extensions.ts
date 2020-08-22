import { RegistryExtensionFieldsForList } from '../graphql-operations'
import { ExtensionCategory, EXTENSION_CATEGORIES } from '../../../shared/src/schema/extensionSchema'
import {
    ConfiguredRegistryExtension,
    toConfiguredRegistryExtension,
    isExtensionEnabled,
} from '../../../shared/src/extensions/extension'
import { validCategories } from './extension/extension'
import { isErrorLike, ErrorLike } from '../../../shared/src/util/errors'
import { ExtensionsEnablement } from './ExtensionsList'
import { Settings, SettingsCascadeOrError } from '../../../shared/src/settings/settings'

export interface CategorizedExtensionRegistry {
    /** Maps categories to ids of extensions  */
    categories: Record<ExtensionCategory, string[]>

    /** All extensions returned by the query indexed by id */
    extensions: { [id: string]: ConfiguredRegistryExtension<RegistryExtensionFieldsForList> }
}

const NO_VALID_CATEGORIES: 'Other'[] = ['Other']

/**
 * Categorizes extensions
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

/**
 *
 *
 * @param categories
 * @param enablement
 */
export function applyExtensionsEnablement(
    categories: CategorizedExtensionRegistry['categories'],
    filteredCategories: ExtensionCategory[],
    enablement: ExtensionsEnablement,
    settings: Settings | ErrorLike | null
): CategorizedExtensionRegistry['categories'] {
    if (enablement === 'all') {
        return categories
    }

    const enabled = enablement === 'enabled'

    return filteredCategories.reduce((toRender, category) => {
        toRender[category] = categories[category].filter(
            extensionID => enabled === isExtensionEnabled(settings, extensionID)
        )
        return toRender
        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    }, {} as Record<ExtensionCategory, string[]>)
}

/**
 * Determines if only default extensions are added
 */
export function onlyDefaultExtensionsAdded(settings: SettingsCascadeOrError): boolean {
    if (!isErrorLike(settings.subjects) && settings.subjects) {
        const userSettings = settings.subjects.find(subject => subject.subject.__typename === 'User')
        const defaultSettings = settings.subjects.find(subject => subject.subject.__typename === 'DefaultSettings')

        if (userSettings && defaultSettings) {
            const userExtensions = !isErrorLike(userSettings.settings) && userSettings.settings?.extensions
            const defaultExtensions = !isErrorLike(defaultSettings.settings) && defaultSettings.settings?.extensions

            if (userExtensions && defaultExtensions) {
                for (const key of Object.keys(userExtensions)) {
                    if (!(key in defaultExtensions)) {
                        return false
                    }
                }
            }
        }
    }

    return true
}
