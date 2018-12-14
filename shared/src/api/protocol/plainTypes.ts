import * as sourcegraph from 'sourcegraph'

/** The plain properties of a {@link module:sourcegraph.WorkspaceRoot}, without methods and accessors. */
export interface WorkspaceRoot {
    uri: string
}

/** The plain properties of a {@link module:sourcegraph.Position}, without methods and accessors. */
export interface Position extends Pick<sourcegraph.Position, 'line' | 'character'> {
    line: number
    character: number
}

/** The plain properties of a {@link module:sourcegraph.Range}, without methods and accessors. */
export interface Range {
    start: Position
    end: Position
}

/** The plain properties of a {@link module:sourcegraph.Selection}, without methods and accessors. */
export interface Selection extends Range {
    isReversed: boolean
}

/** The plain properties of a {@link module:sourcegraph.Location}, without methods and accessors. */
export interface Location {
    uri: string
    range?: Range
}

/** The plain properties of a {@link module:sourcegraph.Definition}, without methods and accessors. */
export type Definition = Location | Location[] | null

/** The plain properties of a {@link module:sourcegraph.Hover}, without methods and accessors. */
export interface Hover extends Pick<sourcegraph.Hover, 'contents' | '__backcompatContents'> {
    range?: Range
}

/** The plain properties of a {@link module:sourcegraph.TextDocumentDecoration}, without methods and accessors. */
export interface TextDocumentDecoration
    extends Pick<sourcegraph.TextDocumentDecoration, Exclude<keyof sourcegraph.TextDocumentDecoration, 'range'>> {
    range: Range
}

export interface Markdown {
    text: string
    html: string
}

export interface SearchResult {
    icon: string
    label: Markdown
    url: string
    detail: Markdown
    matches: SearchMatch[]
}
export interface SearchMatch {
    url: string
    body: Markdown
    highlights: Highlight[]
}
export interface Highlight {
    line: number
    character: number
    length: number
}
