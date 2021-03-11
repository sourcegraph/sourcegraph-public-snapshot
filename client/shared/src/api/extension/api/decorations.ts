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
    // Only need to validate properties that could cause UI errors (e.g. ensure objects aren't passed as React children)
    const validAfter =
        typeof fileDecoration === 'object' &&
        fileDecoration !== null &&
        hasProperty('after')(fileDecoration) &&
        fileDecoration.after &&
        typeof fileDecoration.after === 'object' &&
        hasProperty('contentText')(fileDecoration.after) &&
        typeof fileDecoration.after.contentText === 'string'

    const validMeter =
        typeof fileDecoration === 'object' &&
        fileDecoration !== null &&
        hasProperty('meter')(fileDecoration) &&
        fileDecoration.meter &&
        typeof fileDecoration.meter === 'object' &&
        hasProperty('value')(fileDecoration.meter) &&
        typeof fileDecoration.meter.value === 'number'

    // If neither are valid, no further validation necessary
    if (!(validAfter || validMeter)) {
        return false
    }

    // Check for objects where we expect primitives that will be React children
    const textContentIsObject =
        typeof fileDecoration === 'object' &&
        fileDecoration !== null &&
        hasProperty('after')(fileDecoration) &&
        fileDecoration.after &&
        typeof fileDecoration.after === 'object' &&
        hasProperty('contentText')(fileDecoration.after) &&
        typeof fileDecoration.after.contentText === 'object'

    return !textContentIsObject
}
