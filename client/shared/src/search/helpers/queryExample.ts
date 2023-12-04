import type { FilterType } from '../query/filters'
import { FilterKind, findFilter } from '../query/query'
import type { CharacterRange } from '../query/token'
import { updateFilter } from '../query/transformer'

/**
 * A QueryExample is a structured representation of a query fragment possibly
 * including a placeholder. The token ranges are "adjusted": the meta-characters
 * to indicate the placeholder positions are ignored.
 */
export interface QueryExample {
    tokens: { type: 'text' | 'placeholder'; value: string; start: number; end: number }[]
    // The example value without placeholder meta-characters
    value: string
}

/**
 * Parses the provided example string into a QueryExample. The section enclosed
 * by {...} is interpreted as the placeholder value.
 */
export function createQueryExampleFromString(value: string): QueryExample {
    return parse(value)
}

/**
 * This helper function will update the current query with the provided filter
 * example, updating the existing filter if necessary, and returns the position
 * of the filter and possible placeholder.
 *
 * @param query current search query
 * @param example the filter example to add to the query
 * @param options.singular if true the filter should only occur once in the query
 * @param options.negate if true negate the filter. It's on the caller to verify
 * that the filter is actually negatable.
 * @param options.emptyValue if true the filter will be added or updated with an
 * empty value
 */
export function updateQueryWithFilterAndExample(
    query: string,
    filter: FilterType,
    example: QueryExample,
    { singular = false, negate = false, emptyValue = false } = {}
): { query: string; filterRange: CharacterRange; placeholderRange: CharacterRange } {
    let placeholderRange: CharacterRange | undefined
    let filterRange: CharacterRange
    let field: string = filter

    if (negate) {
        field = '-' + field
    }

    const existingFilter = findFilter(query, field, FilterKind.Global)

    if (existingFilter && singular) {
        // The filter should only appear once. Clear the existing value if
        // necessary.
        if (emptyValue) {
            query = updateFilter(query, field, '')
        }

        const placeholderRangeStart = existingFilter.value?.range.start || existingFilter.field.range.end + 1 // +1 to account for the ":" after the field
        placeholderRange = {
            start: placeholderRangeStart,
            end: emptyValue ? placeholderRangeStart : existingFilter.range.end,
        }

        filterRange = {
            start: existingFilter.range.start,
            end: emptyValue ? placeholderRange.end : existingFilter.range.end,
        }
    } else {
        // Filter can appear multiple times or doesn't exist yet. Always append.
        query = query.trimEnd()

        let rangeStart = query.length
        if (query.length > 0) {
            query += ' '
            rangeStart += 1 // +1 to account for the whitespace that is inserted before the new filter
        }
        query = `${query}${field}:${emptyValue ? '' : example.value}`

        const valueRangeStart = rangeStart + field.length + 1 // +1 to account for the ":" after the field
        if (emptyValue) {
            placeholderRange = {
                start: valueRangeStart,
                end: valueRangeStart,
            }
        } else {
            placeholderRange = getPlaceholderRange(example)
            // Adjust placeholder range to take full query into account
            placeholderRange.start += valueRangeStart
            placeholderRange.end += valueRangeStart
        }

        filterRange = {
            start: rangeStart,
            end: query.length,
        }
    }

    return {
        query,
        placeholderRange,
        filterRange,
    }
}

/**
 * Given a QueryExample object, this function returns the range of the
 * placeholder in the string.
 */
function getPlaceholderRange(example: QueryExample): { start: number; end: number } {
    const token = example.tokens.find(token => token.type === 'placeholder')
    if (!token) {
        throw new Error('Search reference does not contain placeholder.')
    }
    return { start: token.start, end: token.end }
}

/**
 * Advance the index until the provided character is found or the end of the
 * string is reached.
 */
function consumeTillCharacter(input: string, char: string, start: number): number {
    let index = start
    while (input[index] !== char && index < input.length) {
        index++
    }
    return index
}

/**
 * Parse an example value into a sequence of tokens. {...} indicate a placeholder
 * sequence.
 */
function parse(value: string): { tokens: QueryExample['tokens']; value: string } {
    const tokens: QueryExample['tokens'] = []
    let index = 0
    let positionOffset = 0
    let state: 'value' | 'text' = 'text'
    while (index < value.length) {
        switch (state) {
            case 'text': {
                const start = index
                index = consumeTillCharacter(value, '{', index)
                if (index > start) {
                    // exclude empty strings
                    tokens.push({
                        type: 'text',
                        value: value.slice(start, index),
                        start: start - positionOffset,
                        end: index - positionOffset,
                    })
                }
                state = 'value'
                break
            }
            case 'value': {
                // skip {
                index++
                positionOffset++
                const start = index
                index = consumeTillCharacter(value, '}', index)
                if (index === value.length) {
                    throw new Error(`Missing '}' in example value: ${value}`)
                }
                tokens.push({
                    type: 'placeholder',
                    value: value.slice(start, index),
                    start: start - positionOffset,
                    end: index - positionOffset,
                })
                // skip }
                index++
                positionOffset++
                state = 'text'
                break
            }
        }
    }
    return {
        tokens,
        value: tokens.map(token => token.value).join(''),
    }
}
