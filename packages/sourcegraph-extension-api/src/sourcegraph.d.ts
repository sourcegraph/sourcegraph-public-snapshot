/**
 * ## The Sourcegraph extension API
 *
 * Sourcegraph extensions enhance your code host, code reviews, and Sourcegraph itself by adding features such as:
 * - Code intelligence (go-to-definition, find references, hovers, etc.)
 * - Test coverage overlays
 * - Links to live traces, log output, and performance data for a line of code
 * - Git blame
 * - Usage examples for functions
 *
 * Check out the [extension authoring documentation](https://docs.sourcegraph.com/extensions/authoring) to get started.
 */
declare module 'sourcegraph' {
    // tslint:disable member-access

    export interface Unsubscribable {
        unsubscribe(): void
    }

    /**
     * @deprecated In the future the API will use the [native `URL` API](https://developer.mozilla.org/en-US/docs/Web/API/URL)
     */
    export class URI {
        /**
         * @deprecated In the future the API will use the [native `URL` API](https://developer.mozilla.org/en-US/docs/Web/API/URL), which does not contain this method.
         */
        static parse(value: string): URI

        /**
         * @deprecated In the future the API will use the [native `URL` API](https://developer.mozilla.org/en-US/docs/Web/API/URL), which does not contain this method.
         */
        static file(path: string): URI

        constructor(value: string)

        toString(): string

        /**
         * Returns a JSON representation of this Uri.
         *
         * @deprecated In the future the API will use the [native `URL` API](https://developer.mozilla.org/en-US/docs/Web/API/URL), which returns a string instead of an object.
         */
        toJSON(): any
    }

    export class Position {
        /** Zero-based line number. */
        readonly line: number

        /** Zero-based character on the line. */
        readonly character: number

        /**
         * Constructs a Position from a line and character.
         *
         * @param line A zero-based line value.
         * @param character A zero-based character value.
         */
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

    /**
     * A text document, such as a file in a repository.
     */
    export interface TextDocument {
        /**
         * The URI of the text document.
         */
        readonly uri: string

        /**
         * The language of the text document.
         */
        readonly languageId: string

        /**
         * The text contents of the text document.
         */
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
     * Options for an input box displayed as a result of calling {@link Window#showInputBox}.
     */
    export interface InputBoxOptions {
        /**
         * The text that describes what input the user should provide.
         */
        prompt?: string

        /**
         * The pre-filled input value for the input box.
         */
        value?: string
    }

    export interface ProgressOptions {
        title?: string
    }

    export interface Progress {
        /** Optional message. If not set, the previous message is still shown. */
        message?: string

        /** Integer from 0 to 100. If not set, the previous percentage is still shown. */
        percentage?: number
    }

    export interface ProgressReporter {
        /**
         * Updates the progress display with a new message and/or percentage.
         */
        next(status: Progress): void

        /**
         * Turns the progress display into an error display for the given error or message.
         * Use if the operation failed.
         * No further progress updates can be sent after this.
         */
        error(error: any): void

        /**
         * Completes the progress bar and hides the display.
         * Sending a percentage of 100 has the same effect.
         * No further progress updates can be sent after this.
         */
        complete(): void
    }

    /**
     * A window in the client application that is running the extension.
     */
    export interface Window {
        /**
         * The user interface view components that are visible in the window.
         */
        visibleViewComponents: ViewComponent[]

        /**
         * The currently active view component in the window.
         */
        activeViewComponent: ViewComponent | undefined

        /**
         * Show a notification message to the user that does not require interaction or steal focus.
         *
         * @deprecated This API will change.
         * @param message The message to show. Markdown is supported.
         * @return A promise that resolves when the user dismisses the message.
         */
        showNotification(message: string): void

        /**
         * Show progress in the window. Progress is shown while running the given callback
         * and while the promise it returned isn't resolved nor rejected.
         *
         * @param task A callback returning a promise. Progress state can be reported with
         * the provided [ProgressReporter](#ProgressReporter)-object.
         *
         * @return The Promise the task-callback returned.
         */
        withProgress<R>(options: ProgressOptions, task: (reporter: ProgressReporter) => Promise<R>): Promise<R>

        /**
         * Show progress in the window. The returned ProgressReporter can be used to update the
         * progress bar, complete it or turn the notification into an error notification in case the operation failed.
         *
         * @return A ProgressReporter that allows updating the progress display.
         */
        showProgress(options: ProgressOptions): Promise<ProgressReporter>

        /**
         * Show a modal message to the user that the user must dismiss before continuing.
         *
         * @param message The message to show.
         * @return A promise that resolves when the user dismisses the message.
         */
        showMessage(message: string): Promise<void>

        /**
         * Displays an input box to ask the user for input.
         *
         * The returned value will be `undefined` if the input box was canceled (e.g., because the user pressed the
         * ESC key). Otherwise the returned value will be the string provided by the user.
         *
         * @param options Configures the behavior of the input box.
         * @return The string provided by the user, or `undefined` if the input box was canceled.
         */
        showInputBox(options?: InputBoxOptions): Promise<string | undefined>
    }

    /**
     * A user interface component in an application window.
     *
     * Each {@link ViewComponent} has a distinct {@link ViewComponent#type} value that indicates what kind of
     * component it is ({@link CodeEditor}, etc.).
     */
    export type ViewComponent = CodeEditor

    /**
     * A style for a {@link TextDocumentDecoration}.
     */
    export interface ThemableDecorationStyle {
        /** The CSS background-color property value for the line. */
        backgroundColor?: string

        /** The CSS border property value for the line. */
        border?: string

        /** The CSS border-color property value for the line. */
        borderColor?: string

        /** The CSS border-width property value for the line. */
        borderWidth?: string
    }

    /**
     * A text document decoration changes the appearance of a range in the document and/or adds other content to
     * it.
     */
    export interface TextDocumentDecoration extends ThemableDecorationStyle {
        /**
         * The range that the decoration applies to. Currently, decorations are
         * only applied only on the start line, and the entire line. Multiline
         * and intra-line ranges are not supported.
         */
        range: Range

        /**
         * If true, the decoration applies to all lines in the range (inclusive), even if not all characters on the
         * line are included.
         */
        isWholeLine?: boolean

        /** Content to display after the range. */
        after?: DecorationAttachmentRenderOptions

        /** Overwrite style for light themes. */
        light?: ThemableDecorationStyle

        /** Overwrite style for dark themes. */
        dark?: ThemableDecorationStyle
    }

    /**
     * A style for {@link DecorationAttachmentRenderOptions}.
     */
    export interface ThemableDecorationAttachmentStyle {
        /** The CSS background-color property value for the attachment. */
        backgroundColor?: string

        /** The CSS color property value for the attachment. */
        color?: string
    }

    /** A decoration attachment adds content after a {@link TextDocumentDecoration}. */
    export interface DecorationAttachmentRenderOptions extends ThemableDecorationAttachmentStyle {
        /** Text to display in the attachment. */
        contentText?: string

        /** Tooltip text to display when hovering over the attachment. */
        hoverMessage?: string

        /** If set, the attachment becomes a link with this destination URL. */
        linkURL?: string

        /** Overwrite style for light themes. */
        light?: ThemableDecorationAttachmentStyle

        /** Overwrite style for dark themes. */
        dark?: ThemableDecorationAttachmentStyle
    }

    /**
     * Represents a handle to a set of decorations.
     *
     * To get an instance of {@link TextDocumentDecorationType}, use {@link createDecorationType}
     */
    export interface TextDocumentDecorationType {
        key: string
    }

    /**
     * A text editor for code files (as opposed to a rich text editor for documents or other kinds of file format
     * editors).
     */
    export interface CodeEditor {
        /** The type tag for this kind of {@link ViewComponent}. */
        type: 'CodeEditor'

        /**
         * The text document that is open in this editor. The document remains the same for the entire lifetime of
         * this editor.
         */
        readonly document: TextDocument

        /**
         * The primary selection in this text editor. This is equivalent to `CodeEditor.selections[0] || null`.
         *
         * @todo Make this non-readonly.
         */
        readonly selection: Selection | null

        /**
         * The selections in this text editor. A text editor has zero or more selections. The primary selection
         * ({@link CodeEditor#selection}), if any selections exist, is always at index 0.
         *
         * @todo Make this non-readonly.
         */
        readonly selections: Selection[]

        /**
         * Add a set of decorations to this editor. If a set of decorations already exists with the given
         * {@link DecorationType}, they will be replaced.
         *
         * @see {@link TextDocumentDecorationType}
         *
         */
        setDecorations(decorationType: TextDocumentDecorationType, decorations: TextDocumentDecoration[]): void
    }

    /**
     * A panel view created by {@link sourcegraph.app.createPanelView}.
     */
    export interface PanelView extends Unsubscribable {
        /**
         * The title of the panel view.
         */
        title: string

        /**
         * The content to show in the panel view. Markdown is supported.
         */
        content: string

        /**
         * The priority of this panel view. A higher value means that the item is shown near the beginning (usually
         * the left side).
         */
        priority: number

        /**
         * Display the results of the location provider (with the given ID) in this panel below the
         * {@link PanelView#contents}.
         *
         * @internal Experimental. Subject to change or removal without notice.
         */
        component: { locationProvider: string } | null
    }

    /**
     * The client application that is running the extension.
     */
    export namespace app {
        /**
         * The currently active window, or `undefined`. The active window is the window that has focus, or when
         * none has focus, the window that was most recently focused.
         */
        export const activeWindow: Window | undefined

        /**
         * All application windows that are accessible by the extension.
         *
         * @readonly
         */
        export const windows: Window[]

        /**
         * Create a panel view for the view contribution with the given {@link id}.
         *
         * @todo Consider requiring extensions to specify these statically in package.json's contributions section
         * to improve the activation experience.
         *
         * @param id The ID of the view. This may be shown to the user (e.g., in the URL fragment when the panel is
         * active).
         * @returns The panel view.
         */
        export function createPanelView(id: string): PanelView

        /**
         * Creates a decorationType that can be used to add decorations to code views.
         *
         * Use this to create a unique handle to a set of decorations, that can be applied to
         * text editors using {@link setDecorations}.
         */
        export function createDecorationType(): TextDocumentDecorationType
    }

    /**
     * A workspace root is a directory that has been added to a workspace. A workspace can have zero or more roots.
     * Often, each root is the root directory of a repository.
     */
    export interface WorkspaceRoot {
        /**
         * The URI of the root.
         *
         * @todo The format of this URI will be changed in the future. It must not be relied on.
         *
         * @example git://github.com/sourcegraph/sourcegraph?sha#mydir1/mydir2
         */
        readonly uri: URI
    }

    /**
     * The logical workspace that the extension is running in, which may consist of multiple folders, projects, and
     * repositories.
     */
    export namespace workspace {
        /**
         * All text documents currently known to the system.
         *
         * @readonly
         */
        export const textDocuments: TextDocument[]

        /**
         * An event that is fired when a new text document is opened.
         */
        export const onDidOpenTextDocument: Subscribable<TextDocument>

        /**
         * The root directories of the workspace, if any.
         *
         * @example The repository that is currently being viewed is a root.
         * @todo Currently only a single root is supported.
         * @readonly
         */
        export const roots: ReadonlyArray<WorkspaceRoot>

        /**
         * An event that is fired when a workspace root is added or removed from the workspace.
         */
        export const onDidChangeRoots: Subscribable<void>
    }

    /**
     * The full configuration value, containing all settings for the current subject.
     *
     * @template C The configuration schema.
     */
    export interface Configuration<C extends object> {
        /**
         * Returns a value at a specific key in the configuration.
         *
         * @template C The configuration schema.
         * @template K Valid key on the configuration object.
         * @param key The name of the configuration property to get.
         * @return The configuration value, or `undefined`.
         */
        get<K extends keyof C>(key: K): Readonly<C[K]> | undefined

        /**
         * Updates the configuration value for the given key. The updated configuration value is persisted by the
         * client.
         *
         * @template C The configuration schema.
         * @template K Valid key on the configuration object.
         * @param key The name of the configuration property to update.
         * @param value The new value, or undefined to remove it.
         * @return A promise that resolves when the client acknowledges the update.
         */
        update<K extends keyof C>(key: K, value: C[K] | undefined): Promise<void>

        /**
         * The configuration value as a plain object.
         */
        readonly value: Readonly<C>
    }

    /**
     * The configuration settings.
     *
     * It may be merged from the following sources of settings, in order:
     *
     * Default settings
     * Global settings
     * Organization settings (for all organizations the user is a member of)
     * User settings
     * Repository settings
     * Directory settings
     *
     * @todo Add a way to get/update configuration for a specific scope or subject.
     * @todo Support applying defaults to the configuration values.
     */
    export namespace configuration {
        /**
         * Returns the full configuration object.
         *
         * @template C The configuration schema.
         * @return The full configuration object.
         */
        export function get<C extends object = { [key: string]: any }>(): Configuration<C>

        /**
         * Subscribe to changes to the configuration. The {@link next} callback is called when any configuration
         * value changes (and synchronously immediately). Call {@link get} in the callback to obtain the new
         * configuration values.
         *
         * @template C The configuration schema.
         * @return An unsubscribable to stop calling the callback for configuration changes.
         */
        export function subscribe(next: () => void): Unsubscribable
    }

    /**
     * A provider result represents the values that a provider, such as the {@link HoverProvider}, may return. The
     * result may be a single value, a Promise that resolves to a single value, or a Subscribable that emits zero
     * or more values.
     */
    export type ProviderResult<T> =
        | T
        | undefined
        | null
        | Promise<T | undefined | null>
        | Subscribable<T | undefined | null>

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
         * @default MarkupKind.Markdown
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

        /** @deprecated */
        __backcompatContents?: (MarkupContent | string | { language: string; value: string })[]

        /**
         * The range to which this hover applies. When missing, the editor will use the range at the current
         * position or the current position itself.
         */
        range?: Range

        /**
         * The priority for this hover, which determines whether it is shown and the order in which
         * it is shown. Hovers with higher priorities are shown first.
         *
         * If undefined, the hover is always shown. If this hover's priority is a negative number
         * and there is another hover for the same position with a greater priority, this hover is
         * not shown.
         *
         * For example, "fuzzy" or "imprecise" hovers can have negative priorities to ensure they
         * will not be displayed if there is a more precise hover available.
         */
        priority?: number
    }

    export interface HoverProvider {
        provideHover(document: TextDocument, position: Position): ProviderResult<Hover>
    }

    /**
     * The definition of a symbol represented as one or many [locations](#Location). For most programming languages
     * there is only one location at which a symbol is defined. If no definition can be found `null` is returned.
     */
    export type Definition = Location | Location[] | null

    /**
     * A definition provider implements the "go-to-definition" feature.
     */
    export interface DefinitionProvider {
        /**
         * Provide the definition of the symbol at the given position and document.
         *
         * @param document The document in which the command was invoked.
         * @param position The position at which the command was invoked.
         * @return A definition location, or an array of definitions, or `null` if there is no definition.
         */
        provideDefinition(document: TextDocument, position: Position): ProviderResult<Definition>
    }

    /**
     * A type definition provider implements the "go-to-type-definition" feature.
     *
     * @deprecated Use {@link LocationProvider} and {@link sourcegraph.languages.registerLocationProvider} instead.
     */
    export interface TypeDefinitionProvider {
        /**
         * Provide the type definition of the symbol at the given position and document.
         *
         * @deprecated Use {@link LocationProvider} and {@link sourcegraph.languages.registerLocationProvider}
         * instead.
         * @param document The document in which the command was invoked.
         * @param position The position at which the command was invoked.
         * @return A type definition location, or an array of definitions, or `null` if there is no type
         *         definition.
         */
        provideTypeDefinition(document: TextDocument, position: Position): ProviderResult<Definition>
    }

    /**
     * An implementation provider implements the "go-to-implementations" and "go-to-interfaces" features.
     *
     * @deprecated Use {@link LocationProvider} and {@link sourcegraph.languages.registerLocationProvider} instead.
     */
    export interface ImplementationProvider {
        /**
         * Provide the implementations of the symbol at the given position and document.
         *
         * @deprecated Use {@link LocationProvider} and {@link sourcegraph.languages.registerLocationProvider}
         * instead.
         * @param document The document in which the command was invoked.
         * @param position The position at which the command was invoked.
         * @return Implementation locations, or `null` if there are none.
         */
        provideImplementation(document: TextDocument, position: Position): ProviderResult<Definition>
    }

    /**
     * Additional information and parameters for a references request.
     */
    export interface ReferenceContext {
        /** Include the declaration of the current symbol. */
        includeDeclaration: boolean
    }

    /**
     * The reference provider interface defines the contract between extensions and
     * the [find references](https://code.visualstudio.com/docs/editor/editingevolved#_peek)-feature.
     */
    export interface ReferenceProvider {
        /**
         * Provides a set of workspace-wide references for the given position in a document.
         *
         * @param document The document in which the command was invoked.
         * @param position The position at which the command was invoked.
         * @param context Additional information and parameters for the request.
         * @return An array of reference locations.
         */
        provideReferences(
            document: TextDocument,
            position: Position,
            context: ReferenceContext
        ): ProviderResult<Location[]>
    }

    /**
     * A location provider implements features such as "find implementations" and "find type definition". It is the
     * general form of {@link DefinitionProvider} and {@link ReferenceProvider}.
     */
    export interface LocationProvider {
        /**
         * Provide related locations for the symbol at the given position and document.
         *
         * @param document The document in which the command was invoked.
         * @param position The position at which the command was invoked.
         * @return Related locations, or `null` if there are none.
         */
        provideLocations(document: TextDocument, position: Position): ProviderResult<Location[]>
    }

    export namespace languages {
        /**
         * Registers a hover provider, which returns a formatted hover message (intended for display in a tooltip)
         * when the user hovers on code.
         *
         * Multiple providers can be registered for a language. In that case, providers are queried in parallel and
         * the results are merged. A failing provider (rejected promise or exception) will not cause the whole
         * operation to fail.
         *
         * @param selector A selector that defines the documents this provider is applicable to.
         * @param provider A hover provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerHoverProvider(selector: DocumentSelector, provider: HoverProvider): Unsubscribable

        /**
         * Registers a definition provider.
         *
         * Multiple providers can be registered for a language. In that case, providers are queried in parallel and
         * the results are merged. A failing provider (rejected promise or exception) will not cause the whole
         * operation to fail.
         *
         * @param selector A selector that defines the documents this provider is applicable to.
         * @param provider A definition provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerDefinitionProvider(
            selector: DocumentSelector,
            provider: DefinitionProvider
        ): Unsubscribable

        /**
         * Registers a type definition provider.
         *
         * Multiple providers can be registered for a language. In that case, providers are queried in parallel and
         * the results are merged. A failing provider (rejected promise or exception) will not cause the whole
         * operation to fail.
         *
         * @deprecated Use {@link LocationProvider} and {@link registerLocationProvider} instead.
         * @param selector A selector that defines the documents this provider is applicable to.
         * @param provider A type definition provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerTypeDefinitionProvider(
            selector: DocumentSelector,
            provider: TypeDefinitionProvider
        ): Unsubscribable

        /**
         * Registers an implementation provider.
         *
         * Multiple providers can be registered for a language. In that case, providers are queried in parallel and
         * the results are merged. A failing provider (rejected promise or exception) will not cause the whole
         * operation to fail.
         *
         * @deprecated Use {@link LocationProvider} and {@link registerLocationProvider} instead.
         * @param selector A selector that defines the documents this provider is applicable to.
         * @param provider An implementation provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerImplementationProvider(
            selector: DocumentSelector,
            provider: ImplementationProvider
        ): Unsubscribable

        /**
         * Registers a reference provider.
         *
         * Multiple providers can be registered for a language. In that case, providers are queried in parallel and
         * the results are merged. A failing provider (rejected promise or exception) will not cause the whole
         * operation to fail.
         *
         * @param selector A selector that defines the documents this provider is applicable to.
         * @param provider A reference provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerReferenceProvider(
            selector: DocumentSelector,
            provider: ReferenceProvider
        ): Unsubscribable

        /**
         * Registers a generic provider of a list of locations. It is the general form of
         * {@link registerDefinitionProvider} and {@link registerReferenceProvider}. It is intended for "find
         * implementations", "find type definition", and other similar features.
         *
         * The provider can be executed with the `executeLocationProvider` builtin command, passing the {@link id}
         * as the first argument. For more information, see
         * https://docs.sourcegraph.com/extensions/authoring/builtin_commands#executeLocationProvider.
         *
         * @param id An identifier for this location provider that distinguishes it from other location providers.
         * @param selector A selector that defines the documents this provider is applicable to.
         * @param provider A location provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerLocationProvider(
            id: string,
            selector: DocumentSelector,
            provider: LocationProvider
        ): Unsubscribable
    }

    /**
     * A query transformer alters a user's search query before executing a search.
     *
     * Query transformers allow extensions to define new search query operators and syntax, for example,
     * by matching strings in a query (e.g. `go.imports:`) and replacing them with a regular expression or string.
     */
    export interface QueryTransformer {
        /**
         * Transforms a search query into another, valid query. If there are no transformations to be made
         * the original query is returned.
         *
         * @param query A search query.
         */
        transformQuery(query: string): string | Promise<string>
    }

    /**
     * API for extensions to augment search functionality.
     */
    export namespace search {
        /**
         * Registers a query transformer.
         *
         * Multiple transformers can be registered. In that case, all transformations will be applied
         * and the result is a single query that has been altered by all transformers. The order in
         * which transfomers are applied is not defined.
         *
         * @param provider A query transformer.
         */
        export function registerQueryTransformer(provider: QueryTransformer): Unsubscribable
    }

    /**
     * Commands are functions that are implemented and registered by extensions. Extensions can invoke any command
     * (including commands registered by other extensions). The extension can also define contributions (in
     * package.json), such as actions and menu items, that invoke a command.
     */
    export namespace commands {
        /**
         * Registers a command that can be invoked by an action or menu item, or directly (with
         * {@link commands.executeCommand}).
         *
         * @param command A unique identifier for the command.
         * @param callback A command function. If it returns a {@link Promise}, execution waits until it is
         *                 resolved.
         * @return Unsubscribable to unregister this command.
         * @throws Registering a command with an existing command identifier throws an error.
         */
        export function registerCommand(command: string, callback: (...args: any[]) => any): Unsubscribable

        /**
         * Executes the command with the given command identifier.
         *
         * @template T The result type of the command.
         * @param command Identifier of the command to execute.
         * @param rest Parameters passed to the command function.
         * @return A {@link Promise} that resolves to the result of the given command.
         * @throws If no command exists wih the given command identifier, an error is thrown.
         */
        export function executeCommand<T = any>(command: string, ...args: any[]): Promise<T>
    }

    export interface ContextValues {
        [key: string]: string | number | boolean | null
    }

    /**
     * Internal API for Sourcegraph extensions. Most of these will be removed for the beta release of Sourcegraph
     * extensions. They are necessary now due to limitations in the extension API and its implementation that will
     * be addressed in the beta release.
     *
     * @internal
     * @hidden
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
         * Updates context values for use in context expressions and contribution labels.
         *
         * @param updates The updates to apply to the context. If a context property's value is null, it is deleted from the context.
         */
        export function updateContext(updates: ContextValues): void

        /**
         * The URL to the Sourcegraph site that the user's session is associated with. This refers to
         * Sourcegraph.com (`https://sourcegraph.com`) by default, or a self-hosted instance of Sourcegraph.
         *
         * @todo Consider removing this when https://github.com/sourcegraph/sourcegraph/issues/566 is fixed.
         *
         * @example `https://sourcegraph.com`
         */
        export const sourcegraphURL: URI

        /**
         * The client application that is running this extension, either 'sourcegraph' for Sourcegraph or 'other'
         * for all other applications (such as GitHub, GitLab, etc.).
         *
         * @todo Consider removing this when https://github.com/sourcegraph/sourcegraph/issues/566 is fixed.
         */
        export const clientApplication: 'sourcegraph' | 'other'
    }

    /** Support types for {@link Subscribable}. */
    interface NextObserver<T> {
        closed?: boolean
        next: (value: T) => void
        error?: (err: any) => void
        complete?: () => void
    }
    interface ErrorObserver<T> {
        closed?: boolean
        next?: (value: T) => void
        error: (err: any) => void
        complete?: () => void
    }
    interface CompletionObserver<T> {
        closed?: boolean
        next?: (value: T) => void
        error?: (err: any) => void
        complete: () => void
    }
    type PartialObserver<T> = NextObserver<T> | ErrorObserver<T> | CompletionObserver<T>

    /**
     * A stream of values that may be subscribed to.
     */
    export interface Subscribable<T> {
        /**
         * Subscribes to the stream of values.
         *
         * @returns An unsubscribable that, when its {@link Unsubscribable#unsubscribe} method is called, causes
         * the subscription to stop reacting to the stream.
         */
        subscribe(observer?: PartialObserver<T>): Unsubscribable
        subscribe(next?: (value: T) => void, error?: (error: any) => void, complete?: () => void): Unsubscribable
    }

    /**
     * The extension context is passed to the extension's activate function and contains utilities for the
     * extension lifecycle.
     *
     * @since Sourcegraph 3.0-preview. Use `export function activate(ctx?: ExtensionContext) { ... }` for prior
     * versions (to ensure your code handles the pre-3.0-preview case when `ctx` is undefined).
     */
    export interface ExtensionContext {
        /**
         * An object that maintains subscriptions to resources that should be freed when the extension is
         * deactivated.
         *
         * When an extension is deactivated, first its exported `deactivate` function is called (if one exists).
         * The `deactivate` function may be async, in which case deactivation blocks on it finishing. Next,
         * regardless of whether the `deactivate` function finished successfully or rejected with an error, all
         * unsubscribables passed to {@link ExtensionContext#subscriptions#add} are unsubscribed from.
         *
         * (An extension is deactivated when the user disables it, or after an arbitrary time period if its
         * activationEvents no longer evaluate to true.)
         */
        subscriptions: {
            /**
             * Mark a resource's teardown function to be called when the extension is deactivated.
             *
             * @param unsubscribable An {@link Unsubscribable} that frees (unsubscribes from) a resource, or a
             * plain function that does the same. Async functions are not supported. (If deactivation requires
             * async operations, make the `deactivate` function async; that is supported.)
             */
            add: (unsubscribable: Unsubscribable | (() => void)) => void
        }
    }
}
