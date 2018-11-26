import { parse, ParseError, ParseErrorCode } from '@sqs/jsonc-parser'
import * as GQL from '../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeOrError } from '../../../shared/src/settings/settings'
import { createAggregateError } from '../../../shared/src/util/errors'

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
