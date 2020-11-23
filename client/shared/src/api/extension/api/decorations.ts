import { uniqueId } from 'lodash'
import { FileDecoration, TextDocumentDecorationType } from 'sourcegraph'
import { hasProperty } from '../../../util/types'

// LINE DECORATIONS

export const createDecorationType = (): TextDocumentDecorationType => ({ key: uniqueId('TextDocumentDecorationType') })

// FILE DECORATIONS

/**
 * Returns whether the given value is a valid file decoration
 */
export function validateFileDecoration(fileDecoration: unknown): fileDecoration is FileDecoration {
    // TODO(tj): Create validators for every provider result to prevent UI errors

    const validText =
        typeof fileDecoration === 'object' &&
        fileDecoration !== null &&
        hasProperty('text')(fileDecoration) &&
        fileDecoration.text &&
        typeof fileDecoration.text === 'object' &&
        hasProperty('value')(fileDecoration.text) &&
        typeof fileDecoration.text.value === 'string'

    const validPercentage =
        typeof fileDecoration === 'object' &&
        fileDecoration !== null &&
        hasProperty('percentage')(fileDecoration) &&
        fileDecoration.percentage &&
        typeof fileDecoration.percentage === 'object' &&
        hasProperty('value')(fileDecoration.percentage) &&
        typeof fileDecoration.percentage.value === 'number'

    return validText || validPercentage
}
