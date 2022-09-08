/**
 * The specification used to provide search-based code navigation for a
 * particular language. This includes things like file extensions, comment
 * patterns and delimiters, and logic for filtering out obviously wrong
 * search results for definitions.
 */
export interface LanguageSpec {
    /**
     * Used to label markdown code blocks.
     */
    languageID: string

    /**
     * Languages that the extension should be activated for in addition to
     * its main language.
     *
     * @example ["javascript"] for the TypeScript extension.
     */
    additionalLanguages?: string[]

    /**
     * The name of the generated extension.
     */
    stylized: string

    /**
     * The part of the filename after the `.` (e.g. `cpp` in `main.cpp`). This
     * used to restrict the definition and reference searches to filenames within
     * the same "ecosystem".
     */
    fileExts: string[]

    /**
     * Non-glob filenames that match this language. For example, ["Dockerfile"].
     */
    verbatimFilenames?: string[]

    /**
     * Regex that matches individual characters in an identifier.
     */
    identCharPattern?: RegExp

    /**
     * Instructions on how to parse comments in order to extract docstrings.
     */
    commentStyles: CommentStyle[]

    /**
     * Callback that filters the given symbol search results (e.g. to drop
     * results from non-imported files).
     */
    filterDefinitions?: FilterDefinitions

    /**
     * Affects messaging about adding LSIF indexing when hovering over symbols
     * from this language.
     */
    lsifSupport?: LSIFSupport

    /**
     * Supports textDocument/implementation Requests
     */
    textDocumentImplemenationSupport?: boolean
}

export enum LSIFSupport {
    None = 'none',
    Experimental = 'experimental',
    Robust = 'robust',
}

/**
 * Comment patterns and delimiters for a particular language.
 */
export interface CommentStyle {
    /**
     * Identifies a single-line comment. Also used to prevent jump-to-definition
     * into comments (except when the token appears to refer to code).
     *
     * Python example: `/#\s?/`
     */
    lineRegex?: RegExp

    /**
     * The style of block comments.
     */
    block?: BlockCommentStyle

    /**
     * Specifies where documentation is placed relative to the definition.
     * Defaults to `'above the definition'`. In Python, documentation is placed
     * `'below the definition'`.
     */
    docPlacement?: TextDocumentPlacement

    /**
     * Regex that matches lines between a definition and the docstring that
     * should be ignored. Java example: `/^\s*@/` for class/method annotations.
     */
    docstringIgnore?: RegExp
}

/**
 * Where a docstring is located relative to a definition.
 */
export type TextDocumentPlacement = 'above the definition' | 'below the definition'

/**
 * Block comment delimiter patterns for a particular language.
 */
export interface BlockCommentStyle {
    /**
     * Matches the start of a block comment. C++ example: `/\/\*\*?/`
     */
    startRegex: RegExp

    /**
     * Matches the end of a block comment. C++ example: `/\*\//`
     */
    endRegex: RegExp

    /**
     * Matches the noise at the beginning of each line in a block comment after
     * the start, end, and leading indentation have been stripped. C++ example:
     * `/(\s\*\s?)?/`
     */
    lineNoiseRegex?: RegExp
}

/**
 * A filter function that prunes imprecise definitions from search results.
 */
export type FilterDefinitions = <T extends Result>(results: T[], context: FilterContext) => T[]

/**
 * Additional context supplied when filtering search results.
 */
export interface FilterContext {
    /**
     * The name of the repository containing of the current file.
     */
    repo: string

    /**
     * The path to the current file relative to the repository root.
     */
    filePath: string

    /**
     * The full text content of the current file.
     */
    fileContent: string
}

/**
 * Result represents a search result returned from the Sourcegraph API.
 */
export interface Result {
    /**
     * The name of the repository containing the result.
     */
    repo: string

    /**
     * The path to the result file relative to the repository root.
     */
    file: string
}
