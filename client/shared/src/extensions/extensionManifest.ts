import { isPlainObject } from 'lodash'

import { ErrorLike, isErrorLike, parseJSONCOrError } from '@sourcegraph/common'

import { ExtensionManifest as ExtensionManifestSchema } from '../schema/extensionSchema'

/**
 * Represents an input object that is validated against a subset of properties of the {@link ExtensionManifest}
 * JSON Schema. For simplicity, only necessary fields are validated and included here.
 */
export type ExtensionManifest = Pick<
    ExtensionManifestSchema,
    'url' | 'icon' | 'iconDark' | 'activationEvents' | 'contributes' | 'publisher'
>

/**
 * Parses and validates the extension manifest. If parsing or validation fails, an error value is returned (not
 * thrown).
 *
 * @todo Contribution ("contributes" property) validation is incomplete.
 */
export function parseExtensionManifestOrError(input: string): ExtensionManifest | ErrorLike {
    const value = parseJSONCOrError<ExtensionManifest>(input)
    if (!isErrorLike(value)) {
        if (!isPlainObject(value)) {
            return new Error('invalid extension manifest: must be a JSON object')
        }
        const problems: string[] = []
        if (!value.url) {
            problems.push('"url" property must be set')
        } else if (typeof value.url !== 'string') {
            problems.push('"url" property must be a string')
        }
        if (!value.activationEvents) {
            problems.push('"activationEvents" property must be set')
        } else if (!Array.isArray(value.activationEvents)) {
            problems.push('"activationEvents" property must be an array')
        } else if (!value.activationEvents.every(event => typeof event === 'string')) {
            problems.push('"activationEvents" property must be an array of strings')
        }
        if (value.contributes) {
            if (!isPlainObject(value.contributes)) {
                problems.push('"contributes" property must be an object')
            }
        }
        if (value.icon && typeof value.icon !== 'string') {
            problems.push('"icon" property must be a string')
        }
        if (value.iconDark && typeof value.iconDark !== 'string') {
            problems.push('"iconDark" property must be a string')
        }
        if (problems.length > 0) {
            return new Error(`invalid extension manifest: ${problems.join(', ')}`)
        }
    }
    return value
}
