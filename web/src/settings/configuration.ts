import { parse, ParseError, ParseErrorCode } from '@sqs/jsonc-parser'
import settingsSchemaJSON from '../../../schema/settings.schema.json'
import { ConfiguredRegistryExtension } from '../../../shared/src/extensions/extension'
import * as GQL from '../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeOrError } from '../../../shared/src/settings/settings'
import { createAggregateError, isErrorLike } from '../../../shared/src/util/errors'

/**
 * Parses the JSON input using the error-tolerant parser used for site config and settings.
 */
export function parseJSON(text: string): any {
    const errors: ParseError[] = []
    const o = parse(text, errors, { allowTrailingComma: true, disallowComments: false })
    if (errors.length > 0) {
        throw createAggregateError(
            errors.map(v => ({
                ...v,
                code: ParseErrorCode[v.error],
                message: `Configuration parse error, code: ${v.error} (offset: ${v.offset}, length: ${v.length})`,
            }))
        )
    }
    return o
}

export function toGQLKeyPath(keyPath: (string | number)[]): GQL.IKeyPathSegment[] {
    return keyPath.map(v => (typeof v === 'string' ? { property: v } : { index: v }))
}

export function getLastIDForSubject(settingsCascade: SettingsCascadeOrError, subject: GQL.ID): number | null {
    if (!isSettingsValid(settingsCascade)) {
        throw new Error('invalid settings')
    }

    // Find the settings lastID so we can update the settings.
    const subjectInfo = settingsCascade.subjects.find(s => s.subject.id === subject)
    if (!subjectInfo) {
        throw new Error('unable to find owner (settings subject) of saved search')
    }
    return subjectInfo.lastID
}

/**
 * Merges settings schemas from base settings and all configured extensions.
 *
 * @param configuredExtensions
 * @returns A JSON Schema that describes an instance of settings for a particular subject.
 */
export function mergeSettingsSchemas(configuredExtensions: Pick<ConfiguredRegistryExtension, 'manifest'>[]): any {
    return {
        allOf: [
            { $ref: settingsSchemaJSON.$id },
            ...(configuredExtensions || [])
                .map(ce => {
                    if (
                        ce.manifest &&
                        !isErrorLike(ce.manifest) &&
                        ce.manifest.contributes &&
                        ce.manifest.contributes.configuration
                    ) {
                        // Adjust the schema to describe a valid instance of settings for a subject (instead of the
                        // final, merged settings).
                        //
                        // This is necessary to avoid erroneous validation errors. For example, suppose an extension's
                        // configuration schema declares that the property "x" is required. For the configuration to be
                        // valid, "x" may be set in global, organization, or user settings. It is valid for user
                        // settings to NOT contain "x" (if global or organization settings contains "x").
                        //
                        // The JSON Schema returned by mergeSettingsSchema is used for a single subject's settings
                        // (e.g., for user settings in the above example). Therefore, we must allow additionalProperties
                        // and set required to [] to avoid erroneous validation errors.
                        return {
                            ...ce.manifest.contributes.configuration,

                            // Force allow additionalProperties to prevent any single extension's configuration schema
                            // from invalidating all other extensions' configuration properties.
                            additionalProperties: true,

                            // Force no required properties because this instance is only the settings for a single
                            // subject. It is possible that a required property is specified at a different subject in
                            // the cascade, in which case we don't want to report this instance as invalid.
                            required: [],
                        }
                    }
                    return true // JSON Schema that matches everything
                })
                .filter(schema => schema !== true), // omit trivial JSON Schemas
        ],
    }
}
