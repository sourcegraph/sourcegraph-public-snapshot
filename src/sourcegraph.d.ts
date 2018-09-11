/**
 * The Sourcegraph extension API.
 *
 * @todo Work in progress.
 */
declare module 'sourcegraph' {
    // tslint:disable member-access

    export class URI {
        static parse(value: string): URI
        static file(path: string): URI

        constructor(value: string)

        toString(): string

        /**
         * Returns a JSON representation of this Uri.
         *
         * @return An object.
         */
        toJSON(): any
    }

    export class Position {
        readonly line: number
        readonly character: number
        constructor(line: number, character: number)

        /**
         * Check if this position is before `other`.
         *
         * @param other A position.
         * @return `true` if position is on a smaller line
         * or on the same line on a smaller character.
         */
        isBefore(other: Position): boolean

        /**
         * Check if this position is before or equal to `other`.
         *
         * @param other A position.
         * @return `true` if position is on a smaller line
         * or on the same line on a smaller or equal character.
         */
        isBeforeOrEqual(other: Position): boolean

        /**
         * Check if this position is after `other`.
         *
         * @param other A position.
         * @return `true` if position is on a greater line
         * or on the same line on a greater character.
         */
        isAfter(other: Position): boolean

        /**
         * Check if this position is after or equal to `other`.
         *
         * @param other A position.
         * @return `true` if position is on a greater line
         * or on the same line on a greater or equal character.
         */
        isAfterOrEqual(other: Position): boolean

        /**
         * Check if this position is equal to `other`.
         *
         * @param other A position.
         * @return `true` if the line and character of the given position are equal to
         * the line and character of this position.
         */
        isEqual(other: Position): boolean

        /**
         * Compare this to `other`.
         *
         * @param other A position.
         * @return A number smaller than zero if this position is before the given position,
         * a number greater than zero if this position is after the given position, or zero when
         * this and the given position are equal.
         */
        compareTo(other: Position): number

        /**
         * Create a new position relative to this position.
         *
         * @param lineDelta Delta value for the line value, default is `0`.
         * @param characterDelta Delta value for the character value, default is `0`.
         * @return A position which line and character is the sum of the current line and
         * character and the corresponding deltas.
         */
        translate(lineDelta?: number, characterDelta?: number): Position

        /**
         * Derived a new position relative to this position.
         *
         * @param change An object that describes a delta to this position.
         * @return A position that reflects the given delta. Will return `this` position if the change
         * is not changing anything.
         */
        translate(change: { lineDelta?: number; characterDelta?: number }): Position

        /**
         * Create a new position derived from this position.
         *
         * @param line Value that should be used as line value, default is the [existing value](#Position.line)
         * @param character Value that should be used as character value, default is the [existing value](#Position.character)
         * @return A position where line and character are replaced by the given values.
         */
        with(line?: number, character?: number): Position

        /**
         * Derived a new position from this position.
         *
         * @param change An object that describes a change to this position.
         * @return A position that reflects the given change. Will return `this` position if the change
         * is not changing anything.
         */
        with(change: { line?: number; character?: number }): Position
    }

    /**
     * A range represents an ordered pair of two positions.
     * It is guaranteed that [start](#Range.start).isBeforeOrEqual([end](#Range.end))
     *
     * Range objects are __immutable__. Use the [with](#Range.with),
     * [intersection](#Range.intersection), or [union](#Range.union) methods
     * to derive new ranges from an existing range.
     */
    export class Range {
        /**
         * The start position. It is before or equal to [end](#Range.end).
         */
        readonly start: Position

        /**
         * The end position. It is after or equal to [start](#Range.start).
         */
        readonly end: Position

        /**
         * Create a new range from two positions. If `start` is not
         * before or equal to `end`, the values will be swapped.
         *
         * @param start A position.
         * @param end A position.
         */
        constructor(start: Position, end: Position)

        /**
         * Create a new range from number coordinates. It is a shorter equivalent of
         * using `new Range(new Position(startLine, startCharacter), new Position(endLine, endCharacter))`
         *
         * @param startLine A zero-based line value.
         * @param startCharacter A zero-based character value.
         * @param endLine A zero-based line value.
         * @param endCharacter A zero-based character value.
         */
        constructor(startLine: number, startCharacter: number, endLine: number, endCharacter: number)

        /**
         * `true` if `start` and `end` are equal.
         */
        isEmpty: boolean

        /**
         * `true` if `start.line` and `end.line` are equal.
         */
        isSingleLine: boolean

        /**
         * Check if a position or a range is contained in this range.
         *
         * @param positionOrRange A position or a range.
         * @return `true` if the position or range is inside or equal
         * to this range.
         */
        contains(positionOrRange: Position | Range): boolean

        /**
         * Check if `other` equals this range.
         *
         * @param other A range.
         * @return `true` when start and end are [equal](#Position.isEqual) to
         * start and end of this range.
         */
        isEqual(other: Range): boolean

        /**
         * Intersect `range` with this range and returns a new range or `undefined`
         * if the ranges have no overlap.
         *
         * @param range A range.
         * @return A range of the greater start and smaller end positions. Will
         * return undefined when there is no overlap.
         */
        intersection(range: Range): Range | undefined

        /**
         * Compute the union of `other` with this range.
         *
         * @param other A range.
         * @return A range of smaller start position and the greater end position.
         */
        union(other: Range): Range

        /**
         * Derived a new range from this range.
         *
         * @param start A position that should be used as start. The default value is the [current start](#Range.start).
         * @param end A position that should be used as end. The default value is the [current end](#Range.end).
         * @return A range derived from this range with the given start and end position.
         * If start and end are not different `this` range will be returned.
         */
        with(start?: Position, end?: Position): Range

        /**
         * Derived a new range from this range.
         *
         * @param change An object that describes a change to this range.
         * @return A range that reflects the given change. Will return `this` range if the change
         * is not changing anything.
         */
        with(change: { start?: Position; end?: Position }): Range
    }

    /**
     * Represents a text selection in an editor.
     */
    export class Selection extends Range {
        /**
         * The position at which the selection starts.
         * This position might be before or after [active](#Selection.active).
         */
        anchor: Position

        /**
         * The position of the cursor.
         * This position might be before or after [anchor](#Selection.anchor).
         */
        active: Position

        /**
         * Create a selection from two positions.
         *
         * @param anchor A position.
         * @param active A position.
         */
        constructor(anchor: Position, active: Position)

        /**
         * Create a selection from four coordinates.
         *
         * @param anchorLine A zero-based line value.
         * @param anchorCharacter A zero-based character value.
         * @param activeLine A zero-based line value.
         * @param activeCharacter A zero-based character value.
         */
        constructor(anchorLine: number, anchorCharacter: number, activeLine: number, activeCharacter: number)

        /**
         * A selection is reversed if [active](#Selection.active).isBefore([anchor](#Selection.anchor)).
         */
        isReversed: boolean
    }

    /**
     * Represents a location inside a resource, such as a line
     * inside a text file.
     */
    export class Location {
        /**
         * The resource identifier of this location.
         */
        uri: URI

        /**
         * The document range of this location.
         */
        range?: Range

        /**
         * Creates a new location object.
         *
         * @param uri The resource identifier.
         * @param rangeOrPosition The range or position. Positions will be converted to an empty range.
         */
        constructor(uri: URI, rangeOrPosition?: Range | Position)
    }

    export interface TextDocument {
        readonly uri: string
        readonly languageId: string
        readonly text: string
    }

    /**
     * A document filter denotes a document by different properties like the
     * [language](#TextDocument.languageId), the scheme of its resource, or a glob-pattern that is
     * applied to the [path](#TextDocument.fileName).
     *
     * @sample A language filter that applies to typescript files on disk: `{ language: 'typescript', scheme: 'file' }`
     * @sample A language filter that applies to all package.json paths: `{ language: 'json', pattern: '**package.json' }`
     */
    export type DocumentFilter =
        | {
              /** A language id, such as `typescript`. */
              language: string
              /** A URI scheme, such as `file` or `untitled`. */
              scheme?: string
              /** A glob pattern, such as `*.{ts,js}`. */
              pattern?: string
          }
        | {
              /** A language id, such as `typescript`. */
              language?: string
              /** A URI scheme, such as `file` or `untitled`. */
              scheme: string
              /** A glob pattern, such as `*.{ts,js}`. */
              pattern?: string
          }
        | {
              /** A language id, such as `typescript`. */
              language?: string
              /** A URI scheme, such as `file` or `untitled`. */
              scheme?: string
              /** A glob pattern, such as `*.{ts,js}`. */
              pattern: string
          }

    /**
     * A document selector is the combination of one or many document filters.
     *
     * @sample `let sel: DocumentSelector = [{ language: 'typescript' }, { language: 'json', pattern: '**∕tsconfig.json' }]`;
     */
    export type DocumentSelector = (string | DocumentFilter)[]

    /**
     * A provider result represents the values a provider, like the [`HoverProvider`](#HoverProvider),
     * may return. For once this is the actual result type `T`, like `Hover`, or a thenable that resolves
     * to that type `T`. In addition, `null` and `undefined` can be returned - either directly or from a
     * thenable.
     *
     * The snippets below are all valid implementations of the [`HoverProvider`](#HoverProvider):
     *
     * ```ts
     * let a: HoverProvider = {
     * 	provideHover(doc, pos, token): ProviderResult<Hover> {
     * 		return new Hover('Hello World');
     * 	}
     * }
     *
     * let b: HoverProvider = {
     * 	provideHover(doc, pos, token): ProviderResult<Hover> {
     * 		return new Promise(resolve => {
     * 			resolve(new Hover('Hello World'));
     * 	 	});
     * 	}
     * }
     *
     * let c: HoverProvider = {
     * 	provideHover(doc, pos, token): ProviderResult<Hover> {
     * 		return; // undefined
     * 	}
     * }
     * ```
     */
    export type ProviderResult<T> = T | undefined | null | Promise<T | undefined | null>

    /**
     * The definition of a symbol represented as one or many [locations](#Location). For most programming languages
     * there is only one location at which a symbol is defined. If no definition can be found `null` is returned.
     */
    export type Definition = Location | Location[] | null

    /** The kinds of markup that can be used. */
    export const enum MarkupKind {
        PlainText = 'plaintext',
        Markdown = 'markdown',
    }

    /**
     * Human-readable text that supports various kinds of formatting.
     */
    export interface MarkupContent {
        /** The marked up text. */
        value: string

        /**
         * The kind of markup used.
         *
         * @default MarkupKind.PlainText
         */
        kind?: MarkupKind
    }

    /**
     * A hover represents additional information for a symbol or word. Hovers are rendered in a tooltip-like
     * widget.
     */
    export interface Hover {
        /**
         * The contents of this hover.
         */
        contents: MarkupContent

        /**
         * The range to which this hover applies. When missing, the editor will use the range at the current
         * position or the current position itself.
         */
        range?: Range
    }

    export interface HoverProvider {
        provideHover(document: TextDocument, position: Position): ProviderResult<Hover>
    }

    export function registerHoverProvider(selector: DocumentSelector, provider: HoverProvider): Unsubscribable

    /**
     * Internal API for Sourcegraph extensions. These will be removed for the beta release of
     * Sourcegraph extensions. They are necessary now due to limitations in the extension API and
     * its implementation that will be addressed in the beta release.
     *
     * @internal
     */
    export namespace internal {
        /**
         * Returns a promise that resolves when all pending messages have been sent to the client.
         * It helps enforce serialization of messages.
         *
         * @internal
         */
        export function sync(): Promise<void>

        /**
         * Information about the client's capabilities beyond that which is handled by the extension
         * host.
         *
         * @internal
         */
        export const experimentalCapabilities: any

        /**
         * The underlying connection to the Sourcegraph extension client.
         *
         * @deprecated
         * @internal
         */
        export const rawConnection: any
    }

    export interface Unsubscribable {
        unsubscribe(): void
    }
}
