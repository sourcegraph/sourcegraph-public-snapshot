import * as Monaco from 'monaco-editor'
import { Token, toMonacoRange } from './scanner'
import { resolveFilter } from './filters'
import { decorateTokens, DecoratedToken, RegexpMetaKind } from './tokens'

const toHover = (token: DecoratedToken): string => {
    switch (token.type) {
        case 'pattern': {
            const quantity = token.value.length > 1 ? 'string' : 'character'
            return `Matches the ${quantity} \`${token.value}\`.`
        }
        case 'regexpMeta': {
            switch (token.kind) {
                case RegexpMetaKind.Alternative:
                    return '**Or**. Match either the expression before or after the `|`.'
                case RegexpMetaKind.Assertion:
                    switch (token.value) {
                        case '^':
                            return '**Start anchor**. Match the beginning of a string. Typically used to match a string prefix, as in `^prefix`. Also often used with the end anchor `$` to match an exact string, as in `^exact$`.'
                        case '$':
                            return '**End anchor**. Match the end of a string. Typically used to match a string suffix, as in `suffix$`. Also often used with the start anchor to match an exact string, as in `^exact$`.'
                        case '\\b':
                            return '**Word boundary**. Match a position where a word character comes after a non-word character, or vice versa. Typically used to match whole words, as in `\\bword\\b`.'
                        case '\\B':
                            return '**Negated word boundary**. Match a position between two word characters, or a position between two non-word characters. This is the negation of `\\b`.'
                    }
                case RegexpMetaKind.CharacterClass:
                    return token.value.startsWith('[^')
                        ? '**Negated character class**. Match any character _not_ inside the square brackets.'
                        : '**Character class**. Match any character inside the square brackets.'
                case RegexpMetaKind.CharacterSet:
                    switch (token.value) {
                        case '.':
                            return '**Dot**. Match any character except a line break.'
                        case '\\w':
                            return '**Word**. Match any word character. '
                        case '\\W':
                            return '**Negated word**. Match any non-word character. Matches any character that is **not** an alphabetic character, digit, or underscore.'
                        case '\\d':
                            return '**Digit**. Match any digit character `0-9`.'
                        case '\\D':
                            return '**Negated digit**. Match any character that is **not** a digit `0-9`.'
                        case '\\s':
                            return '**Whitespace**. Match any whitespace character like a space, line break, or tab.'
                        case '\\S':
                            return '**Negated whitespace**. Match any character that is **not** a whitespace character like a space, line break, or tab.'
                    }
                case RegexpMetaKind.Delimited:
                    return '**Group**. Groups together multiple expressions to match.'
                case RegexpMetaKind.EscapedCharacter: {
                    const escapable = '~`!@#$%^&*()[]{}<>,.?/\\|=+-_'
                    let description = escapable.includes(token.value[1])
                        ? `Match the character \`${token.value[1]}\`.`
                        : `The character \`${token.value[1]}\` is escaped.`
                    switch (token.value[1]) {
                        case 'n':
                            description = 'Match a new line.'
                            break
                        case 't':
                            description = 'Match a tab.'
                            break
                        case 'r':
                            description = 'Match a carriage return.'
                            break
                    }
                    return `**Escaped Character. ${description}`
                }
                case RegexpMetaKind.LazyQuantifier:
                    return '**Lazy**. Match as few as characters as possible that match the previous expression.'
                case RegexpMetaKind.RangeQuantifier:
                    switch (token.value) {
                        case '*':
                            return '**Zero or more**. Match zero or more of the previous expression.'
                        case '?':
                            return '**Optional**. Match zero or one of the previous expression.'
                        case '+':
                            return '**One or more**. Match one or more of the previous expression.'
                        default: {
                            const range = token.value.slice(1, -1).split(',')
                            let quantity = ''
                            if (range.length === 1 || (range.length === 2 && range[0] === range[1])) {
                                quantity = range[0]
                            } else if (range[1] === '') {
                                quantity = `${range[0]} or more`
                            } else {
                                quantity = `between ${range[0]} and ${range[1]}`
                            }
                            return `**Range**. Match ${quantity} of the previous expression.`
                        }
                    }
            }
        }
    }
    return ''
}

const inside = (column: number) => ({ range }: Pick<Token | DecoratedToken, 'range'>): boolean =>
    range.start + 1 <= column && range.end >= column

/**
 * Returns the hover result for a hovered search token in the Monaco query input.
 */
export const getHoverResult = (
    tokens: Token[],
    { column }: Pick<Monaco.Position, 'column'>,
    smartQuery = false
): Monaco.languages.Hover | null => {
    const tokensAtCursor = (smartQuery ? decorateTokens(tokens) : tokens).filter(inside(column))
    if (tokensAtCursor.length === 0) {
        return null
    }
    const values: string[] = []
    let range: Monaco.IRange | undefined
    tokensAtCursor.map(token => {
        switch (token.type) {
            case 'filter': {
                // This 'filter' branch only exists to preserve previous behavior when smmartQuery is false.
                // When smartQuery is true, 'filter' tokens are handled by the 'field' case and its values in
                // the rest of this switch statement.
                const resolvedFilter = resolveFilter(token.field.value)
                if (resolvedFilter) {
                    values.push(
                        'negated' in resolvedFilter
                            ? resolvedFilter.definition.description(resolvedFilter.negated)
                            : resolvedFilter.definition.description
                    )
                    range = toMonacoRange(token.range)
                }
                break
            }
            case 'field': {
                const resolvedFilter = resolveFilter(token.value)
                if (resolvedFilter) {
                    values.push(
                        'negated' in resolvedFilter
                            ? resolvedFilter.definition.description(resolvedFilter.negated)
                            : resolvedFilter.definition.description
                    )
                    // Add 1 to end of range to include the ':'.
                    range = toMonacoRange({ start: token.range.start, end: token.range.end + 1 })
                }
                break
            }
            case 'pattern':
                values.push(toHover(token))
                range = toMonacoRange(token.range)
                break
            case 'regexpMeta':
                values.push(toHover(token))
                range = toMonacoRange(token.groupRange ? token.groupRange : token.range)
                break
        }
    })
    return {
        contents: values.map<Monaco.IMarkdownString>(
            (value): Monaco.IMarkdownString => ({
                value,
            })
        ),
        range,
    }
}
