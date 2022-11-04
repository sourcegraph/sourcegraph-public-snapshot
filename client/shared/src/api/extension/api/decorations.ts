import { uniqueId } from 'lodash'
import {
    TextDocumentDecorationType,
    DecorationAttachmentRenderOptions,
    ThemableDecorationAttachmentStyle,
    ThemableDecorationStyle,
    FileDecoration,
} from 'sourcegraph'

import { hasProperty } from '@sourcegraph/common'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'

// LINE DECORATIONS

export const createDecorationType = (): TextDocumentDecorationType => ({ key: uniqueId('TextDocumentDecorationType') })

/**
 * Resolves the actual styles to use for the attachment based on the current theme.
 */
export function decorationStyleForTheme(
    attachment: TextDocumentDecoration,
    isLightTheme: boolean
): ThemableDecorationStyle {
    const overrides = isLightTheme ? attachment.light : attachment.dark
    // Discard non-ThemableDecorationStyle properties so they aren't included in result.
    const { range, isWholeLine, after, light, dark, ...base } = attachment
    return { ...base, ...overrides }
}

/**
 * Resolves the actual styles to use for the attachment based on the current theme.
 */
export function decorationAttachmentStyleForTheme(
    attachment: DecorationAttachmentRenderOptions,
    isLightTheme: boolean
): ThemableDecorationAttachmentStyle {
    const overrides = isLightTheme ? attachment.light : attachment.dark
    // Discard non-ThemableDecorationAttachmentStyle properties so they aren't included in result.
    const { contentText, hoverMessage, linkURL, light, dark, ...base } = attachment
    return { ...base, ...overrides }
}

export type DecorationMapByLine = ReadonlyMap<number, TextDocumentDecoration[]>

/**
 * @returns Map from 1-based line number to non-empty array of TextDocumentDecoration for that line
 *
 * @todo this does not handle decorations that span multiple lines
 */
export const groupDecorationsByLine = (decorations: TextDocumentDecoration[] | null): DecorationMapByLine => {
    const grouped = new Map<number, TextDocumentDecoration[]>()
    for (const decoration of decorations || []) {
        const lineNumber = decoration.range.start.line + 1
        const decorationsForLine = grouped.get(lineNumber)
        if (!decorationsForLine) {
            grouped.set(lineNumber, [decoration])
        } else {
            decorationsForLine.push(decoration)
        }
    }
    return grouped
}

// FILE DECORATIONS

/**
 * Resolves the actual color to use for the file decoration based on the current theme and selection state.
 */
export function fileDecorationColorForTheme(
    after: NonNullable<FileDecoration['after']>,
    isLightTheme: boolean,
    isActive?: boolean
): string | undefined {
    const overrides = isLightTheme ? after.light : after.dark
    // Discard non-ThemableFileDecorationStyle properties so they aren't included in result.
    const { light, dark, contentText, hoverMessage, ...base } = after
    const merged = { ...base, ...overrides }

    // fall back to default color if active color isn't specified
    return isActive ? merged.activeColor ?? merged.color : merged.color
}

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
