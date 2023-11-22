import {
    modify as jsoncModify,
    applyEdits,
    type JSONPath,
    type FormattingOptions,
    parse,
    type ParseError,
    format as jsoncFormat,
    printParseErrorCode,
} from 'jsonc-parser'

import { asError, createAggregateError, type ErrorLike } from '../errors'

/**
 * Parses the JSON input using an error-tolerant "JSONC" parser. If an error occurs, it is returned as a value
 * instead of being thrown. This is useful when input is parsed in the background (not in response to any specific
 * user action), because it makes it easier to save the error and only show it to the user when it matters (for
 * some interactive user action).
 */
export function parseJSONCOrError<T>(input: string): T | ErrorLike {
    try {
        return parseJSON(input) as T
    } catch (error) {
        return asError(error)
    }
}

/**
 * Parses the JSON input using an error-tolerant "JSONC" parser.
 */
function parseJSON(input: string): any {
    const errors: ParseError[] = []
    const parsed = parse(input, errors, { allowTrailingComma: true, disallowComments: false })
    if (errors.length > 0) {
        throw createAggregateError(
            errors.map(error => ({
                ...error,
                code: error.error,
                message: `parse error (code: ${error.error}, error: ${printParseErrorCode(error.error)}, offset: ${
                    error.offset
                }, length: ${error.length})`,
            }))
        )
    }
    return parsed
}

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

/**
 * Simplified jsonc API method to modify jsonc object.
 * @param originalContent Original content (settings)
 * @param path - path to the field which will be modified
 * @param value - new value for modify field
 */
export const modify = (originalContent: string, path: JSONPath, value: unknown): string => {
    const addingExtensionKeyEdits = jsoncModify(originalContent, path, value, {
        formattingOptions: defaultFormattingOptions,
    })

    return applyEdits(originalContent, addingExtensionKeyEdits)
}

/**
 * Format string with jsonc default format options.
 */
export const format = (content: string): string => {
    const formatEdits = jsoncFormat(content, { offset: 0, length: content.length }, defaultFormattingOptions)

    return applyEdits(content, formatEdits)
}

export const stringify = (object: object | null): string => format(JSON.stringify(object))
