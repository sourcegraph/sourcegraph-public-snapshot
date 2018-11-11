import { parse, ParseError, ParseErrorCode } from '@sqs/jsonc-parser'
import { Observable, ReplaySubject } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphqlschema'
import { Settings } from '../schema/settings.schema'
import { createAggregateError } from '../util/errors'

/**
 * Represents the settings from various subjects from GraphQL (user, orgs, and global).
 */
export const settingsCascade = new ReplaySubject<GQL.ISettingsCascade>(1)

/**
 * Always represents the final settings for the current user or visitor.
 */
export const viewerSettings: Observable<Settings> = settingsCascade.pipe(
    map(cascade => parseJSON(cascade.final) as Settings)
)

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
