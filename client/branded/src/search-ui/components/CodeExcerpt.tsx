import React, { useCallback, useLayoutEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { BehaviorSubject } from 'rxjs'

import { highlightNodeMultiline } from '@sourcegraph/common'
import type { Repo } from '@sourcegraph/shared/src/util/url'
import { Code } from '@sourcegraph/wildcard'

import styles from './CodeExcerpt.module.scss'

interface Props extends Repo {
    commitID: string
    filePath: string
    highlightRanges: HighlightRange[]
    /** The 0-based (inclusive) line number that this code excerpt starts at */
    startLine: number
    /** The 0-based (exclusive) line number that this code excerpt ends at */
    endLine: number
    className?: string

    plaintextLines: string[]
    highlightedLines?: string[]

    onCopy?: () => void
}

export interface HighlightRange {
    /**
     * The 0-based line number where this highlight range begins
     */
    startLine: number
    /**
     * The 0-based character offset from the beginning of startLine where this highlight range begins
     */
    startCharacter: number
    /**
     * The 0-based line number where this highlight range ends
     */
    endLine: number
    /**
     * The 0-based character offset from the beginning of endLine where this highlight range ends
     */
    endCharacter: number
}

/**
 * A code excerpt that displays syntax highlighting and match range highlighting.
 */
export const CodeExcerpt: React.FunctionComponent<Props> = ({
    plaintextLines,
    highlightedLines,
    startLine,
    endLine,
    highlightRanges,
    className,
    onCopy,
}) => {
    const table = highlightedLines ? (
        <table dangerouslySetInnerHTML={{ __html: highlightedLines.join('') }} />
    ) : (
        <table>
            <tbody>
                {plaintextLines.map((line, i) => (
                    <tr key={line}>
                        <td className="line" data-line={startLine + i + 1} />
                        <td className="code">
                            <span className="hl-text hl-plain">{line}</span>
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>
    )

    // Both the behavior subject and the React state are needed here. The behavior subject is
    // used for hoverified events while the React state is used for match highlighting.
    // The state is needed because React won't re-render when the behavior subject's value changes.
    const tableContainerElements = useMemo(() => new BehaviorSubject<HTMLElement | null>(null), [])
    const [tableContainerElement, setTableContainerElement] = useState<HTMLElement | null>(null)
    const updateTableContainerElementReference = useCallback(
        (reference: HTMLElement | null): void => {
            tableContainerElements.next(reference)
            setTableContainerElement(reference)
        },
        [tableContainerElements]
    )

    // Highlight the search matches
    useLayoutEffect(() => {
        if (tableContainerElement) {
            const visibleRows = tableContainerElement.querySelectorAll<HTMLTableRowElement>('table tr')
            for (const highlight of highlightRanges) {
                // Select the HTML rows in the excerpt that correspond to the first and last line to be highlighted.
                // highlight.startLine is the 0-indexed line number in the code file, and startLine is the 0-indexed
                // line number of the first visible line in the excerpt. So, subtract startLine
                // from highlight.startLine to get the correct 0-based index in visibleRows that holds the HTML row
                // where highlighting should begin. Subtract startLine from highlight.endLine to get the correct 0-based
                // index in visibleRows that holds the HTML row where highlighting should end.
                const startRowIndex = highlight.startLine - startLine
                const endRowIndex = highlight.endLine - startLine
                const startRow = visibleRows[startRowIndex]
                const endRow = visibleRows[endRowIndex]
                if (startRow && endRow) {
                    highlightNodeMultiline(
                        visibleRows,
                        startRow,
                        endRow,
                        startRowIndex,
                        endRowIndex,
                        highlight.startCharacter,
                        highlight.endCharacter
                    )
                }
            }
        }
    }, [highlightRanges, startLine, endLine, tableContainerElement])

    return (
        <Code data-testid="code-excerpt" onCopy={onCopy} className={classNames(styles.codeExcerpt, className)}>
            <div ref={updateTableContainerElementReference}>{table}</div>
        </Code>
    )
}
