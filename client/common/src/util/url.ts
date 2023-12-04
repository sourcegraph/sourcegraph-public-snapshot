import type { Position, Range, Selection } from '@sourcegraph/extension-api-types'

import { tryCatch } from '../errors'

/**
 * Provide one.
 * @param position either 1-indexed partial position
 * @param range or 1-indexed partial range spec
 */
export function toPositionOrRangeQueryParameter(context: {
    position?: { line: number; character?: number }
    range?: { start: { line: number; character?: number }; end: { line: number; character?: number } }
}): string | undefined {
    if (context.range) {
        const emptyRange =
            context.range.start.line === context.range.end.line &&
            context.range.start.character === context.range.end.character
        return (
            'L' +
            (emptyRange
                ? toPositionHashComponent(context.range.start)
                : `${toPositionHashComponent(context.range.start)}-${toPositionHashComponent(context.range.end)}`)
        )
    }
    if (context.position) {
        return 'L' + toPositionHashComponent(context.position)
    }
    return undefined
}

/**
 * @param ctx 1-indexed partial position
 */
export function toPositionHashComponent(position: { line: number; character?: number }): string {
    return position.line.toString() + (position.character ? `:${position.character}` : '')
}

/**
 * Represents a line, a position, a line range, or a position range. It forbids
 * just a character, or a range from a line to a position or vice versa (such as
 * "L1-2:3" or "L1:2-3"), none of which would make much sense.
 *
 * 1-indexed.
 */
export type LineOrPositionOrRange =
    | { line?: undefined; character?: undefined; endLine?: undefined; endCharacter?: undefined }
    | { line: number; character?: number; endLine?: undefined; endCharacter?: undefined }
    | { line: number; character?: undefined; endLine?: number; endCharacter?: undefined }
    | { line: number; character: number; endLine: number; endCharacter: number }

export function lprToRange(lpr: LineOrPositionOrRange): Range | undefined {
    if (lpr.line === undefined) {
        return undefined
    }
    return {
        start: { line: lpr.line, character: lpr.character || 0 },
        end: {
            line: lpr.endLine || lpr.line,
            character: lpr.endCharacter || lpr.character || 0,
        },
    }
}

// `lprToRange` sets character to 0 if it's undefined. Only - 1 the character if it's not 0.
const characterZeroIndexed = (character: number): number => (character === 0 ? character : character - 1)

export function lprToSelectionsZeroIndexed(lpr: LineOrPositionOrRange): Selection[] {
    const range = lprToRange(lpr)
    if (range === undefined) {
        return []
    }
    const start: Position = { line: range.start.line - 1, character: characterZeroIndexed(range.start.character) }
    const end: Position = { line: range.end.line - 1, character: characterZeroIndexed(range.end.character) }
    return [
        {
            start,
            end,
            anchor: start,
            active: end,
            isReversed: false,
        },
    ]
}

/**
 * Finds an existing line range search parameter like "L1-2:3"
 */
export function findLineKeyInSearchParameters(searchParameters: URLSearchParams): string | undefined {
    for (const key of searchParameters.keys()) {
        if (key.startsWith('L')) {
            return key
        }
        break
    }
    return undefined
}

/**
 * Encodes revision with encodeURIComponent, except that slashes ('/') are preserved,
 * because they are not ambiguous in any of the current places where used, and URLs
 * for (e.g.) branches with slashes look a lot nicer with '/' than '%2F'.
 */
export function escapeRevspecForURL(revision: string): string {
    return encodeURIPathComponent(revision)
}

export function toViewStateHash(viewState: string | undefined): string {
    return viewState ? `#tab=${viewState}` : ''
}

/**
 * %-Encodes a path component of a URI.
 *
 * It encodes all special characters except forward slashes and the plus sign `+`. The plus sign only has meaning
 * as a space in the query component of a URL, because its special meaning is defined for the
 * `application/x-www-form-urlencoded` MIME type, which is used for queries. It is not part of the general
 * `%`-encoding for URLs.
 */
export const encodeURIPathComponent = (component: string): string =>
    component.split('/').map(encodeURIComponent).join('/').replaceAll('%2B', '+')

/**
 * Returns true if the given URL points outside the current site.
 */
export const isExternalLink = (
    url: string,
    windowLocation__testingOnly: Pick<URL, 'origin' | 'href'> = window.location
): boolean =>
    !!tryCatch(() => new URL(url, windowLocation__testingOnly.href).origin !== windowLocation__testingOnly.origin)

/**
 * Stringifies the provided search parameters, replaces encoded `/` and `:` characters,
 * and removes trailing `=`.
 */
export const formatSearchParameters = (searchParameters: URLSearchParams): string =>
    searchParameters.toString().replaceAll('%2F', '/').replaceAll('%3A', ':').replaceAll('=&', '&').replace(/=$/, '')

export const addLineRangeQueryParameter = (
    searchParameters: URLSearchParams,
    range: string | undefined
): URLSearchParams => {
    const existingLineRangeKey = findLineKeyInSearchParameters(searchParameters)
    if (existingLineRangeKey) {
        searchParameters.delete(existingLineRangeKey)
    }
    // If a non-empty range exists add it to the start of the parameters, otherwise return the existing search parameters
    return range ? new URLSearchParams([[range, ''], ...searchParameters.entries()]) : searchParameters
}

export const appendLineRangeQueryParameter = (url: string, range: string | undefined): string => {
    const newUrl = new URL(url, window.location.href)
    const searchQuery = formatSearchParameters(addLineRangeQueryParameter(newUrl.searchParams, range))
    return newUrl.pathname + `?${searchQuery}` + newUrl.hash
}
