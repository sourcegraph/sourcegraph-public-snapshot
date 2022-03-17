import { Contributions, Raw } from '@sourcegraph/client-api'

/**
 * See the extensions.schema.json JSON Schema for canonical documentation on these types.
 *
 * This file is derived from the extensions.schema.json JSON Schema. It must be updated manually when the JSON
 * Schema or any of its referenced schemas change.
 *
 * TODO: Make this auto-generated. json2ts does not handle the "$ref" well, so it was simpler and faster to just
 * manually duplicate it for now.
 */

/**
 * The set of known categories in the extension registry.
 *
 * Keep this in sync with <extension.schema.json>'s #/categories/items/enum set.
 *
 * This uses a typed array instead of a TypeScript enum to avoid needing to define redundant identifiers for each
 * string constant (e.g., `ProgrammingLanguages = 'Programming languages'`).
 */
export const EXTENSION_CATEGORIES = array([
    'External services',
    'Reports and stats',
    'Linters',
    'Code editors',
    'Code analysis',
    'Programming languages',
    'Other',
])

/**
 * The set of known categories in the extension registry.
 */
export type ExtensionCategory = typeof EXTENSION_CATEGORIES[number]

/**
 * The set of valid extension card header colors for the extension registry.
 *
 * Keep this in sync with <extension.schema.json>'s #/headerColor/enum set.
 */
export const EXTENSION_HEADER_COLORS = new Set([
    'blue',
    'yellow',
    'orange',
    'pink',
    'red',
    'purple',
    'green',
    'grey',
] as const)

/**
 * The set of valid extension card header colors for the extension registry.
 */
export type ExtensionHeaderColor = typeof EXTENSION_HEADER_COLORS extends Set<infer T> ? T : never

export interface ExtensionManifest {
    description?: string
    readme?: string
    url: string
    wip?: boolean
    repository?: {
        type?: string
        url: string
    }

    /**
     * The element type includes `string` because this value has not been validated. Use {@link knownCategories} to
     * filter to only known {@link ExtensionCategory} values.
     */
    categories?: (ExtensionCategory | string)[]

    tags?: string[]
    icon?: string
    iconDark?: string
    headerColor?: ExtensionHeaderColor
    activationEvents: string[]
    contributes?: Raw<Contributions> & { configuration?: { [key: string]: any } }
    publisher?: string
}

/** TypeScript helper for making an array type with constant string union elements, not just string[]. */
function array<T extends string>(a: T[]): T[] {
    return a
}
