import type { ErrorLike } from '@sourcegraph/common'
import type { Position, Range } from '@sourcegraph/extension-api-types'

import type { LOADING } from './loading'
import type { HoveredToken } from './tokenPosition'

/**
 * @template C Extra context for the hovered token.
 * @template D The type of the hover content data.
 * @template A The type of an action.
 */
export interface HoverOverlayProps<C extends object, D, A> {
    /** What to show as contents */
    hoverOrError?: typeof LOADING | (HoverAttachment & D) | null | ErrorLike

    /** The position of the tooltip (assigned to `style`) */
    overlayPosition?: { left: number } & ({ top: number } | { bottom: number })

    /**
     * The hovered token (position and word).
     * Used for the Find References buttons and for error messages
     */
    hoveredToken?: HoveredToken & C

    /**
     * Actions to display as buttons or links in the hover.
     */
    actionsOrError?: typeof LOADING | A[] | null | ErrorLike
}

/**
 * Describes the range in the document (usually a token) that the hover is attached to.
 */
export interface HoverAttachment {
    /**
     * The range to which this hover applies. When missing, it will use the range at the current
     * position or the current position itself.
     */
    range?: Range
}

/**
 * Describes a range in the document that should be highlighted.
 */
export interface DocumentHighlight {
    /**
     * The range to be highlighted.
     */
    range: Range
}

/**
 * Reports whether {@link value} is a {@link Position}.
 */
export function isPosition(value: any): value is Position {
    return value && typeof value.line === 'number' && typeof value.character === 'number'
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
