import * as jsonc from '@sqs/jsonc-parser'

const defaultFormattingOptions: jsonc.FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

/**
 * Simplified jsonc API method to modify jsonc object.
 *
 * @param originalContent Original content (settings)
 * @param path - path to the field which will be modified
 * @param value - new value for modify field
 */
export const modify = (originalContent: string, path: jsonc.JSONPath, value: unknown): string => {
    const addingExtensionKeyEdits = jsonc.modify(originalContent, path, value, {
        formattingOptions: defaultFormattingOptions,
    })

    return jsonc.applyEdits(originalContent, addingExtensionKeyEdits)
}
