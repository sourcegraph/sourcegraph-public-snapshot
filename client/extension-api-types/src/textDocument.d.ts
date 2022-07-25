import * as sourcegraph from 'sourcegraph'

import { Range } from './location'

/**
 * A decoration to apply to a text document.
 *
 * @see module:sourcgraph.TextDocumentDecoration
 */
export interface TextDocumentDecoration
    extends Pick<sourcegraph.TextDocumentDecoration, Exclude<keyof sourcegraph.TextDocumentDecoration, 'range'>> {
    /** The range that the decoration applies to. */
    range: Range
}

export interface InsightDecoration
    extends Pick<sourcegraph.TextDocumentDecoration, Exclude<keyof sourcegraph.TextDocumentDecoration, 'range'>> {
    /** The range that the decoration applies to. */
    range: Range

    /** The raw html to render in line. */
    content: JSX.Element

    /** The JSX Element to render in the popover */
    popover: JSX.Element

    trigger?: 'hover' | 'click'
}
