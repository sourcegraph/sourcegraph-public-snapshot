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
     * @deprecated Use the global [native `URL` API](https://developer.mozilla.org/en-US/docs/Web/API/URL)
     */
    export const URI: typeof URL
    /**
     * @deprecated Use the global [native `URL` API](https://developer.mozilla.org/en-US/docs/Web/API/URL)
     */
    export type URI = URL

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
        readonly uri: URL

        /**
         * The document range of this location.
         */
        readonly range?: Range

        /**
         * Creates a new location object.
         *
         * @param uri The resource identifier.
         * @param rangeOrPosition The range or position. Positions will be converted to an empty range.
         */
        constructor(uri: URL, rangeOrPosition?: Range | Position)
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
         *
         * When using the [Sourcegraph browser
         * extension](https://docs.sourcegraph.com/integration/browser_extension), the value is
         * `undefined` because determining the text contents (in general) is not possible without
         * additional access to the code host API. In the future, this limitation may be removed.
         */
        readonly text: string | undefined

        /**
         * Convert the position to a zero-based offset.
         *
         * The position will be adjusted using {@link TextDocument#validatePosition}.
         *
         * @param position A position.
         * @return A valid zero-based offset.
         * @throws if {@link TextDocument#text} is undefined.
         */
        offsetAt(position: Position): number

        /**
         * Convert a zero-based offset to a position.
         *
         * @param offset A zero-based offset.
         * @return A valid {@link Position}.
         * @throws if {@link TextDocument#text} is undefined.
         */
        positionAt(offset: number): Position

        /**
         * Ensure a position is contained in the range of this document. If not, adjust it so that
         * it is.
         *
         * @param position A position.
         * @return The given position or a new, adjusted position.
         * @throws if {@link TextDocument#text} is undefined.
         */
        validatePosition(position: Position): Position

        /**
         * Ensure a range is completely contained in this document.
         *
         * @param range A range.
         * @return The given range or a new, adjusted range.
         * @throws if {@link TextDocument#text} is undefined.
         */
        validateRange(range: Range): Range

        /**
         * Get the range of the word at the given position.
         *
         * The position will be adjusted using {@link TextDocument#validatePosition}.
         *
         * @param position A position.
         * @return A range spanning a word, or `undefined`.
         */
        getWordRangeAtPosition(position: Position): Range | undefined
    }

    /**
     * A text edit represents edits to apply to a document.
     */
    export class TextEdit {
        /**
         * Utility to create a replace edit.
         *
         * @param range A range.
         * @param newText A string.
         * @return A new text edit object.
         */
        static replace(range: Range, newText: string): TextEdit

        /**
         * Utility to create an insert edit.
         *
         * @param position The insertion position.
         * @param newText A string.
         * @return A new text edit object.
         */
        static insert(position: Position, newText: string): TextEdit

        /**
         * Utility to create a delete edit.
         *
         * @param range A range.
         * @return A new text edit object.
         */
        static delete(range: Range): TextEdit

        /**
         * The range this edit applies to.
         */
        readonly range: Range

        /**
         * The string this edit will insert.
         */
        readonly newText: string

        /**
         * Create a new TextEdit.
         *
         * @param range A range.
         * @param newText A string.
         */
        constructor(range: Range, newText: string)
    }

    /**
     * A workspace edit is a collection of textual and files changes for multiple resources and
     * documents.
     */
    export class WorkspaceEdit {
        /**
         * Get all text edits grouped by resource.
         *
         * @return A shallow copy of `[URL, TextEdit[]]`-tuples.
         */
        textEdits(): IterableIterator<[URL, TextEdit[]]>

        /**
         * Get the text edits for a resource.
         *
         * @param uri A resource identifier.
         * @return An array of text edits.
         */
        get(uri: URL): TextEdit[]

        /**
         * Check if a text edit for a resource exists.
         *
         * @param uri A resource identifier.
         * @return `true` if the given resource will be touched by this edit.
         */
        has(uri: URL): boolean

        /**
         * Set (and replace) text edits for a resource.
         *
         * @param uri A resource identifier.
         * @param edits An array of text edits.
         */
        set(uri: URL, edits: TextEdit[]): void

        /**
         * Replace the given range with given text for the given resource.
         *
         * @param uri A resource identifier.
         * @param range A range.
         * @param newText A string.
         */
        replace(uri: URL, range: Range, newText: string): void

        /**
         * Insert the given text at the given position.
         *
         * @param uri A resource identifier.
         * @param position A position.
         * @param newText A string.
         */
        insert(uri: URL, position: Position, newText: string): void

        /**
         * Delete the text at the given range.
         *
         * @param uri A resource identifier.
         * @param range A range.
         */
        delete(uri: URL, range: Range): void

        /**
         * Create a regular file.
         *
         * @param uri URL of the new file..
         * @param options Defines if an existing file should be overwritten or be
         * ignored. When overwrite and ignoreIfExists are both set overwrite wins.
         */
        createFile(uri: URL, options?: { overwrite?: boolean; ignoreIfExists?: boolean }): void

        /**
         * Delete a file or folder.
         *
         * @param uri The uri of the file that is to be deleted.
         */
        deleteFile(uri: URL, options?: { recursive?: boolean; ignoreIfNotExists?: boolean }): void

        /**
         * Rename a file or folder.
         *
         * @param oldUrl The existing file.
         * @param newUrl The new location.
         * @param options Defines if existing files should be overwritten or be ignored. When
         * overwrite and ignoreIfExists are both set overwrite wins.
         */
        renameFile(oldUrl: URL, newUrl: URL, options?: { overwrite?: boolean; ignoreIfExists?: boolean }): void
    }

    /**
     * A file glob pattern to match file paths against. This can either be a glob pattern string
     * (like `**​/*.{ts,js}` or `*.{ts,js}`) or a [relative pattern](#RelativePattern).
     *
     * Glob patterns can have the following syntax:
     *
     * - `*` to match one or more characters in a path segment
     * - `?` to match on one character in a path segment
     * - `**` to match any number of path segments, including none
     * - `{}` to group conditions (e.g. `**​/*.{ts,js}` matches all TypeScript and JavaScript files)
     * - `[]` to declare a range of characters to match in a path segment (e.g., `example.[0-9]` to
     *   match on `example.0`, `example.1`, …)
     * - `[!...]` to negate a range of characters to match in a path segment (e.g., `example.[!0-9]`
     *   to match on `example.a`, `example.b`, but not `example.0`)
     *
     * @todo Introduce support for VS Code's `RelativePattern`, to make it easy to define globs
     * relative to other globs (and handle things like backslashes correctly).
     */
    export type GlobPattern = string

    /**
     * A document filter denotes a document by different properties like the
     * [language](#TextDocument.languageId), the scheme of its resource, or a glob-pattern that is
     * applied to the [path](#TextDocument.fileName).
     * A document filter matches if all the provided properties (those of `language`, `scheme` and `pattern` that are not `undefined`) match.
     * If all properties are `undefined`, the document filter matches all documents.
     *
     * @sample A language filter that applies to typescript files on disk: `{ language: 'typescript', scheme: 'file' }`
     * @sample A language filter that applies to all package.json paths: `{ language: 'json', pattern: '**package.json' }`
     */
    export interface DocumentFilter {
        /** A language id, such as `typescript` or `*`. */
        language?: string
        /** A URI scheme, such as `file` or `untitled`. */
        scheme?: string
        /** A glob pattern, such as `*.{ts,js}`. */
        pattern?: string
    }

    /**
     * A document selector is the combination of one or many document filters.
     * A document matches the selector if any of the given filters matches.
     * If the filter is a string and not a {@link DocumentFilter}, it will be treated as a language id.
     *
     * @example let sel: DocumentSelector = [{ language: 'typescript' }, { language: 'json', pattern: '**∕tsconfig.json' }];
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
         * An event that is fired when the active view component changes.
         */
        activeViewComponentChanges: Subscribable<ViewComponent | undefined>

        /**
         * Show a notification message to the user that does not require interaction or steal focus.
         *
         * @param message The message to show. Markdown is supported.
         * @param type a {@link NotificationType} affecting the display of the notification.
         */
        showNotification(message: string, type: NotificationType): void

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
     * To get an instance of {@link TextDocumentDecorationType}, use
     * {@link sourcegraph.app.createDecorationType}
     */
    export interface TextDocumentDecorationType {
        /** An opaque identifier. */
        readonly key: string
    }

    /**
     * A text editor for code files (as opposed to a rich text editor for documents or other kinds of file format
     * editors).
     */
    export interface CodeEditor {
        /** The type tag for this kind of {@link ViewComponent}. */
        readonly type: 'CodeEditor'

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
         * An event that is fired when the selections in this text editor change.
         * The primary selection ({@link CodeEditor#selection}), if any selections exist,
         * is always at index 0 of the emitted array.
         */
        readonly selectionsChanges: Subscribable<Selection[]>

        /**
         * Add a set of decorations to this editor. If a set of decorations already exists with the given
         * {@link TextDocumentDecorationType}, they will be replaced.
         *
         * @see {@link TextDocumentDecorationType}
         * @see {@link sourcegraph.app.createDecorationType}
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
         * An event that is fired when the currently active window changes.
         */
        export const activeWindowChanges: Subscribable<Window | undefined>

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
        readonly uri: URL
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
         *
         * @deprecated Renamed to {@link workspace.openedTextDocuments}.
         */
        export const onDidOpenTextDocument: Subscribable<TextDocument>

        /**
         * An event that is fired when a new text document is opened.
         */
        export const openedTextDocuments: Subscribable<TextDocument>

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
         *
         * @deprecated Renamed to {@link workspace.rootsChanges}.
         */
        export const onDidChangeRoots: Subscribable<void>

        /**
         * An event that is fired when a workspace root is added or removed from the workspace.
         */
        export const rootChanges: Subscribable<void>

        /**
         * Find files across all [roots](#workspace.roots) in the workspace.
         *
         * @sample `findFiles('**​/*.js', '**​/node_modules/**', 10)`
         * @param include A [glob pattern](#GlobPattern) that defines the files to search for. The
         * glob pattern will be matched against the file paths of resulting matches relative to
         * their workspace.
         * @param exclude A [glob pattern](#GlobPattern) that defines files and folders to exclude.
         * The glob pattern will be matched against the file paths of resulting matches relative to
         * their workspace. When `undefined` only default excludes will apply, when `null` no
         * excludes will apply.
         * @param maxResults The maximum number of results to return.
         * @return A subscribable that emits once with an array of matching resource URIs and then
         * completes.
         */
        // TODO!(sqs): not implemented, maybe not needed?
        //
        // export function findFiles(
        //     include: GlobPattern,
        //     exclude?: GlobPattern | null,
        //     maxResults?: number
        // ): Subscribable<URL[]>

        /**
         * Opens a document. If this document is not already open when called, the
         * {@link workspace.openedTextDocuments} subscribable will emit this document.
         *
         * @param uri The identifier of the resource to open.
         * @return A promise that resolves to the opened [document](#TextDocument).
         */
        export function openTextDocument(uri: URL): Promise<TextDocument>
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
    export enum MarkupKind {
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
     * The type of a notification shown through {@link Window.showNotification}.
     */
    export enum NotificationType {
        /**
         * An error message.
         */
        Error = 1,
        /**
         * A warning message.
         */
        Warning = 2,
        /**
         * An info message.
         */
        Info = 3,
        /**
         * A log message.
         */
        Log = 4,
        /**
         * A success message.
         */
        Success = 5,
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

    /**
     * A completion item is a suggestion to complete text that the user has typed.
     *
     * @see {@link CompletionItemProvider#provideCompletionItems}
     */
    export interface CompletionItem {
        /**
         * The label of this completion item, which is rendered prominently. If no
         * {@link CompletionItem#insertText} is specified, the label is the text inserted when the
         * user selects this completion.
         */
        label: string

        /**
         * The description of this completion item, which is rendered less prominently but still
         * alongside the {@link CompletionItem#label}.
         */
        description?: string

        /**
         * A string to insert in a document when the user selects this completion. When not set, the
         * {@link CompletionItem#label} is used.
         */
        insertText?: string
    }

    /**
     * A collection of [completion items](#CompletionItem) to be presented in the editor.
     */
    export interface CompletionList {
        /**
         * The list of completions.
         */
        items: CompletionItem[]
    }

    /**
     * A completion item provider provides suggestions to insert or apply at the cursor as the user
     * is typing.
     *
     * Providers are queried for completions as the user types in any document matching the document
     * selector specified at registration time.
     */
    export interface CompletionItemProvider {
        /**
         * Provide completion items for the given position and document.
         *
         * @param document The document in which the command was invoked.
         * @param position The position at which the command was invoked.
         *
         * @return An array of completions, a [completion list](#CompletionList), or a thenable that resolves to either.
         * The lack of a result can be signaled by returning `undefined`, `null`, or an empty array.
         */
        provideCompletionItems(document: TextDocument, position: Position): ProviderResult<CompletionList>
    }

    /**
     * The event that is emitted when diagnostics change.
     */
    export interface DiagnosticChangeEvent {
        /**
         * An array of resources for which diagnostics have changed.
         */
        readonly uris: URL[]
    }

    /**
     * The severity level of a [diagnostic](#Diagnostic).
     */
    export enum DiagnosticSeverity {
        /**
         * Something not allowed by the rules of a language or other means.
         */
        Error = 0,

        /**
         * Something suspicious but allowed.
         */
        Warning = 1,

        /**
         * Something to inform about but not a problem.
         */
        Information = 2,

        /**
         * Something to hint to a better way of doing it, like proposing
         * a refactoring.
         */
        Hint = 3,
    }

    /**
     * Represents a diagnostic, such as a compiler error or warning. Diagnostic objects are only
     * valid in the scope of a file.
     */
    export interface Diagnostic {
        /**
         * The range to which this diagnostic applies.
         */
        readonly range: Range

        /**
         * The human-readable message.
         */
        readonly message: string

        /**
         * The severity, default is [error](#DiagnosticSeverity.Error).
         */
        readonly severity: DiagnosticSeverity

        /**
         * A human-readable string describing the source of this
         * diagnostic, e.g. 'typescript' or 'super lint'.
         */
        readonly source?: string

        /**
         * A code or identifier for this diagnostic. This is useful for, e.g., associating [code
         * actions](#CodeActionContext) with this diagnostic.
         */
        readonly code?: string | number
    }

    /**
     * A diagnostics collection is a container that manages a set of [diagnostics](#Diagnostic).
     * Diagnostics are always scoped to a diagnostics collection and a resource.
     *
     * To get an instance of a {@link DiagnosticCollection}, use
     * [createDiagnosticCollection](#languages.createDiagnosticCollection).
     */
    export interface DiagnosticCollection extends Unsubscribable {
        /**
         * The name of this diagnostic collection, such as `typescript`. Every diagnostic from this
         * collection will be associated with this name.
         */
        readonly name: string

        /**
         * Assign diagnostics for given resource. Existing diagnostics for the resource (if any)
         * will be replaced.
         *
         * @param uri A resource identifier.
         * @param diagnostics Array of diagnostics or `undefined`
         */
        set(uri: URL, diagnostics: Diagnostic[] | undefined): void

        /**
         * Replace all entries in this collection.
         *
         * Diagnostics of multiple tuples of the same uri will be merged, e.g `[[file1, [d1]],
         * [file1, [d2]]]` is equivalent to `[[file1, [d1, d2]]]`. If a diagnostics item is
         * `undefined` as in `[file1, undefined]` all previous but not subsequent diagnostics are
         * removed.
         *
         * @param entries An array of tuples, like `[[file1, [d1, d2]], [file2, [d3, d4, d5]]]`, or
         * `undefined`.
         */
        set(entries: [URL, Diagnostic[] | undefined][]): void

        /**
         * Remove all diagnostics from this collection that belong to the provided `uri`. This is
         * equivalent to `#set(uri, undefined)`.
         *
         * @param uri A resource identifier.
         */
        delete(uri: URL): void

        /**
         * Remove all diagnostics from this collection. The same
         * as calling `#set(undefined)`;
         */
        clear(): void

        /**
         * Get all entries in this collection.
         */
        entries(): IterableIterator<[URL, Diagnostic[]]>

        /**
         * Get the diagnostics for a given resource.
         *
         * @param uri A resource identifier.
         * @returns An immutable array of [diagnostics](#Diagnostic) or `undefined`.
         */
        get(uri: URL): readonly Diagnostic[] | undefined

        /**
         * Check if this collection contains diagnostics for a given resource.
         *
         * @param uri A resource identifier.
         * @returns `true` if this collection has diagnostic for the given resource.
         */
        has(uri: URL): boolean
    }

    /**
     * Contains additional information about the context in which a [code
     * action](#CodeActionProvider.provideCodeActions) is run.
     */
    export interface CodeActionContext {
        /**
         * An array of diagnostics.
         */
        readonly diagnostics: Diagnostic[]
    }

    /**
     * Represents a reference to a command. Provides a title which will be used to represent a
     * command in the UI and, optionally, an array of arguments which will be passed to the command
     * handler function when invoked.
     */
    export interface Command {
        /**
         * Title of the command, such as `Save`.
         */
        title: string

        /**
         * The identifier of the actual command handler.
         * @see [commands.registerCommand](#commands.registerCommand).
         */
        command: string

        /**
         * A tooltip for the command, when represented in the UI.
         */
        tooltip?: string

        /**
         * Arguments that the command handler should be invoked with.
         */
        arguments?: any[]
    }

    /**
     * A code action represents a change that can be performed in code, e.g. to fix a problem or to
     * refactor code.
     *
     * A CodeAction must set either [`edit`](#CodeAction.edit) and/or a
     * [`command`](#CodeAction.command). If both are supplied, the `edit` is applied first, then the
     * command is executed.
     */
    export interface CodeAction {
        /**
         * A short, human-readable, title for this code action.
         */
        readonly title: string

        /**
         * A {@link WorkspaceEdit} that this code action performs.
         */
        readonly edit?: WorkspaceEdit

        /**
         * The [diagnostics](#Diagnostic) that this code action resolves.
         */
        readonly diagnostics?: Diagnostic[]

        /**
         * A [command](#Command) that this code action executes.
         */
        readonly command?: Command
    }

    /**
     * A code action is an action that can be taken on code.
     */
    export interface CodeActionProvider {
        /**
         * Provide code actions for the given document and range.
         *
         * @param document The document to provide code actions for.
         * @param range The selector or range to provide code actions for. This will always be a
         * selection if there is a currently active editor.
         * @param context Context carrying additional information.
         * @return An array of commands, quick fixes, or refactorings or a thenable of such. The
         * lack of a result can be signaled by returning `undefined`, `null`, or an empty array.
         */
        provideCodeActions(
            document: TextDocument,
            range: Range | Selection,
            context: CodeActionContext
        ): ProviderResult<CodeAction[]>
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

        /**
         * Registers a completion item provider.
         *
         * Multiple providers can be registered with overlapping document selectors. In that case,
         * providers are queried in parallel and the results are merged. A failing provider will not
         * cause the whole operation to fail.
         *
         * @param selector A selector that defines the documents this provider applies to.
         * @param provider A completion item provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerCompletionItemProvider(
            selector: DocumentSelector,
            provider: CompletionItemProvider
        ): Unsubscribable

        /**
         * A subscribable that emits when the global set of diagnostics changes.
         */
        export const diagnosticsChanges: Subscribable<DiagnosticChangeEvent>

        /**
         * Get all diagnostics for a given resource.
         *
         * @param resource A resource
         * @returns An array of [diagnostics](#Diagnostic) or an empty array.
         */
        export function getDiagnostics(resource: URL): Diagnostic[]

        /**
         * Get all diagnostics.
         *
         * @returns An array of uri-diagnostics tuples or an empty array.
         */
        export function getDiagnostics(): [URL, Diagnostic[]][]

        /**
         * Create a diagnostics collection.
         *
         * @param name The [name](#DiagnosticCollection.name) of the collection.
         * @return A new diagnostic collection.
         */
        export function createDiagnosticCollection(name: string): DiagnosticCollection

        /**
         * Register a code action provider.
         *
         * Multiple providers can be registered for a language. In that case, providers are queried
         * in parallel and the results are merged. A failing provider (rejected promise or
         * exception) will not cause a failure of the whole operation.
         *
         * @param selector A selector that defines the documents this provider is applicable to.
         * @param provider A code action provider.
         * @return An unsubscribable to unregister this provider.
         */
        export function registerCodeActionProvider(
            selector: DocumentSelector,
            provider: CodeActionProvider
        ): Unsubscribable
    }

    /**
     * The parameters for a search query.
     */
    export interface SearchQuery {
        /**
         * The text pattern to search for.
         */
        pattern: string

        /**
         * The pattern type.
         *
         * @todo Support structural search.
         */
        type: 'regexp'
    }

    /**
     * Include and exclude patterns for searches.
     */
    export interface IncludeExcludePatterns {
        /**
         * Include results whose paths or names match any of these patterns.
         *
         * @todo Allow multiple patterns when globs are supported.
         */
        includes?: [] | [string]

        /**
         * Exclude results whose paths or names match any of these patterns.
         *
         * @todo Allow multiple patterns when globs are supported.
         */
        excludes?: [] | [string]

        /**
         * The pattern type.
         *
         * @todo Support globs.
         */
        type: 'regexp'
    }

    /**
     * Options for searches.
     */
    export interface SearchOptions {
        /**
         * Limit the search to repositories that match these patterns.
         *
         * @todo Support searching one or more repositories at specific non-default-branch revisions
         * (currently only searching repositories' default branch is supported by this API).
         */
        repositories?: IncludeExcludePatterns

        /**
         * Limit the search to files that match these patterns.
         */
        files?: IncludeExcludePatterns

        /** The maximum number of results to return. */
        maxResults?: number
    }

    /**
     * A text search result from {@link search.findTextInFiles}.
     */
    export interface TextSearchResult {
        /** The URI of the matching document. */
        uri: string

        /**
         * The ranges of the match in the document, or undefined if the document was matched based
         * on criteria other than its contents (e.g., based on its filename).
         */
        ranges?: Range[]
    }

    /**
     * A match in a {@link SearchResult} from a {@link SearchResultProvider}.
     */
    export interface SearchResultMatch {
        /**
         * URL to the matched item.
         */
        url: string
        /**
         * A string containing the preview text of the result match, including any context to be displayed.
         * The preview can be any length, and is displayed in its entirety. Can be plain text or Markdown.
         */
        preview: MarkupContent
        /**
         * Highlights are currently only applied if the body is a code block. The highlights
         * are applied after the markdown is rendered; therefore, the line and character count
         * should exclude the markdown code fences.
         */
        highlights: Range[]
    }

    /**
     * A search result from a {@link SearchResultProvider}. A search result may contain multiple matches.
     */
    export interface SearchResult {
        /**
         * URL to an icon to be displayed with each search result.
         * Can be a URL to an image or base-64 encoded data uri.
         */
        iconUrl: string
        /**
         * A prominently-displayed string describing the result. Can be plain text or Markdown.
         */
        label: MarkupContent
        /**
         * URL to the search result. This is a URL to the containing corpus of the {@link SearchResult#matches} (e.g.
         * a file), whereas {@link SearchResultMatch#url} will link to the specific match (e.g. a line in the file).
         * For SearchResults with a single match, this value is often the same as {@link SearchResultMatch#url}.
         *
         * This is used to display a URL associated with search results in text-based clients such as the src-cli.
         */
        url: string
        /**
         * A less prominently-displayed string with secondary information about the result. Can
         * be plain text or Markdown.
         */
        detail: MarkupContent
        /**
         * A list of matches in this search result.
         */
        matches: SearchResultMatch[]
    }

    /**
     * A search result provider accepts a query and returns a list of results.
     * Experimental. Subject to change or removal without notice.
     */
    export interface SearchResultProvider {
        /**
         * Provide results for a search query.
         *
         * @param query A search query.
         */
        provideSearchResults(query: string): ProviderResult<SearchResult[]>
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
     * A provider of text search results.
     */
    export interface TextSearchProvider {
        /**
         * Provide results that match the given query.
         *
         * @param query The query parameters.
         * @param options The search options.
         * @returns A subscribable that emits batches of search results and then completes.
         */
        provideTextSearchResults(query: SearchQuery, options: SearchOptions): Subscribable<TextSearchResult[]>
    }

    /**
     * API for extensions to augment search functionality.
     */
    export namespace search {
        /**
         * Search text in files across all known repositories (including repositories that are not
         * currently open as [workspace roots](#workspace.roots)).
         *
         * @param query The query parameters for the search.
         * @param options The options for the search.
         * @returns A subscribable that emits batches of results and completes when all results have
         * been emitted.
         */
        export function findTextInFiles(query: SearchQuery, options?: SearchOptions): Subscribable<TextSearchResult[]>

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

        /**
         * Register a text search provider.
         *
         * Multiple providers can be registered. In that case, results will be returned grouped by
         * provider. The order in which results from providers are grouped is not defined.
         *
         * @param provider A search result provider.
         * @returns An unsubscribable to unregister the provider.
         */
        export function registerTextSearchProvider(provider: TextSearchProvider): Unsubscribable
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

    /**
     * A description of the information available at a URL.
     */
    export interface LinkPreview {
        /**
         * The content of this link preview, which is shown next to the link.
         */
        content?: MarkupContent

        /**
         * The hover content of this link preview, which is shown when the cursor hovers the link.
         *
         * @todo Add support for Markdown. Currently only plain text is supported.
         */
        hover?: Pick<MarkupContent, 'value'> & { kind: MarkupKind.PlainText }
    }

    /**
     * Called to obtain a preview of the information available at a URL.
     */
    export interface LinkPreviewProvider {
        /**
         * Provides a preview of the information available at the URL of a link in a document.
         *
         * @todo Add a `context` parameter so that the provider knows what document contains the
         * link (so that it can handle links in code files differently from rendered Markdown
         * documents, for example).
         *
         * @param url The URL of the link to preview.
         */
        provideLinkPreview(url: URL): ProviderResult<LinkPreview>
    }

    /**
     * Extensions can customize how content is rendered.
     */
    export namespace content {
        /**
         * EXPERIMENTAL. This API is subject to change without notice and has no compatibility
         * guarantees.
         *
         * Registers a provider for link previews ({@link LinkPreviewProvider}) for all URLs in a
         * document matching the {@link urlMatchPattern}. A link preview is a description of the
         * information available at a URL.
         *
         * @todo Support a more powerful syntax for URL match patterns, such as Chrome's
         * (https://developer.chrome.com/extensions/match_patterns).
         *
         * @param urlMatchPattern A pattern that matches URLs for which the provider is called to
         * obtain a preview. Currently it matches all URLs that start with the match pattern (i.e.,
         * string prefix matches). No wildcards are supported.
         * @param provider The link preview provider.
         */
        export function registerLinkPreviewProvider(
            urlMatchPattern: string,
            provider: LinkPreviewProvider
        ): Unsubscribable
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
        export const sourcegraphURL: URL

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
        /** @deprecated Use an observer instead of a complete callback */
        subscribe(next: null | undefined, error: null | undefined, complete: () => void): Unsubscribable
        /** @deprecated Use an observer instead of an error callback */
        subscribe(next: null | undefined, error: (error: any) => void, complete?: (() => void) | null): Unsubscribable
        /** @deprecated Use an observer instead of a complete callback */
        subscribe(next: (value: T) => void, error: null | undefined, complete: () => void): Unsubscribable
        subscribe(
            next?: ((value: T) => void) | null,
            error?: ((error: any) => void) | null,
            complete?: (() => void) | null
        ): Unsubscribable
    }

    /**
     * The extension context is passed to the extension's activate function and contains utilities for the
     * extension lifecycle.
     *
     * @since Sourcegraph 3.0. Use `export function activate(ctx?: ExtensionContext) { ... }` for prior
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
