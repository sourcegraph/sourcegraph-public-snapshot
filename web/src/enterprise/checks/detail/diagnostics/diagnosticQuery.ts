import * as sourcegraph from 'sourcegraph'

/**
 * The statuses for a diagnostic.
 */
export enum DiagnosticResolutionStatus {
    /** This diagnostic is not fixed and no fix has been selected yet. */
    Unresolved = 'unresolved',

    /** A fix for this diagnostic has been selected but not yet merged. */
    PendingResolution = 'pending',
}

/**
 * A {@link sourcegraph.DiagnosticQuery} with the original query string it was parsed from.
 */
export interface ParsedDiagnosticQuery extends sourcegraph.DiagnosticQuery {
    /**
     * The original query string that the query was parsed from.
     */
    input: string

    /**
     * Only match diagnostics with this resolution status.
     */
    status?: DiagnosticResolutionStatus
}

/**
 * Parse a diagnostic query from its string representation.
 */
export function parseDiagnosticQuery(query: string): ParsedDiagnosticQuery {
    type NonReadonly<T> = { -readonly [P in keyof T]: T[P] }
    const parsed: NonReadonly<ParsedDiagnosticQuery> = { input: query }
    for (const [field, value] of tokenize(query)) {
        switch (field) {
            case 'type:':
                parsed.type = value
                break
            case 'repo:':
                if (!parsed.document) {
                    parsed.document = []
                }
                parsed.document.push({ pattern: `git://${value}**/*` }) // TODO!(sqs)
                break
            case 'is:':
                const status =
                    value === DiagnosticResolutionStatus.Unresolved
                        ? DiagnosticResolutionStatus.Unresolved
                        : value === DiagnosticResolutionStatus.PendingResolution
                        ? DiagnosticResolutionStatus.PendingResolution
                        : undefined
                if (status) {
                    if (parsed.status !== undefined && parsed.status !== status) {
                        delete parsed.status // `is:unresolved is:pending` should match all
                    } else {
                        parsed.status = status
                    }
                }
                break
            case 'tag:':
                if (!parsed.tag) {
                    parsed.tag = []
                }
                parsed.tag.push(value)
                break
            case undefined:
                if (value) {
                    parsed.message = `${parsed.message ? `${parsed.message} ` : ''}${value}`
                }
        }
    }
    return parsed
}

const FIELDS = ['type:', 'tag:', 'is:', 'repo:'] as const

type ParsedToken = [typeof FIELDS[number] | undefined, string]

function parseToken(token: string): ParsedToken {
    for (const field of FIELDS) {
        if (token.startsWith(field)) {
            return [field, token.slice(field.length)]
        }
    }
    return [undefined, token]
}

function* tokenize(query: string): Iterable<ParsedToken> {
    for (const token of query.split(/\s+/g)) {
        yield parseToken(token)
    }
}

/**
 * Append a `field:value` to a diagnostic query and return the new query string.
 */
export function appendToDiagnosticQuery(query: string, field: typeof FIELDS[number], value: string): string {
    return `${query ? `${query} ` : ''}${field}${value}`
}

/**
 * Replace a `field:value` in a diagnostic query and return the new query string.
 */
export function replaceInDiagnosticQuery(
    query: string,
    fieldToReplace: typeof FIELDS[number],
    newValue: string
): string {
    const keepTokens = Array.from(tokenize(query)).filter(([field]) => field !== fieldToReplace)
    keepTokens.push([fieldToReplace, newValue])
    return keepTokens.map(([field, value]) => `${field || ''}${value}`).join(' ')
}

/**
 * Reports whether the `field:value` is present in the diagnostic query.
 */
export function isInDiagnosticQuery(query: string, field: typeof FIELDS[number], value: string): boolean {
    return Array.from(tokenize(query)).some(([aField, aValue]) => field === aField && value === aValue)
}
