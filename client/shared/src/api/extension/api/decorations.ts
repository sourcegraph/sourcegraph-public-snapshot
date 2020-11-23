import { uniqueId } from 'lodash'
import { FileDecoration, TextDocumentDecorationType } from 'sourcegraph'

// LINE DECORATIONS

export const createDecorationType = (): TextDocumentDecorationType => ({ key: uniqueId('TextDocumentDecorationType') })

// FILE DECORATIONS

/**
 * Returns whether the given value is a valid file decoration
 */
export function validateFileDecoration(fileDecoration: FileDecoration): boolean {
    // TODO(tj): Create validators for every provider result to prevent UI errors

    const validText = typeof fileDecoration.text === 'object' && typeof fileDecoration.text.value === 'string'
    const validPercentage =
        typeof fileDecoration.percentage === 'object' && typeof fileDecoration.percentage.value === 'number'

    return validText || validPercentage
}
