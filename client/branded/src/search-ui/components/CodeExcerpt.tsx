import React, { useCallback, useEffect, useLayoutEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import VisibilitySensor from 'react-visibility-sensor'
import { type Observable, type Subscription, BehaviorSubject } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { asError, type ErrorLike, isErrorLike, highlightNodeMultiline } from '@sourcegraph/common'
import type { Repo } from '@sourcegraph/shared/src/util/url'
import { Code } from '@sourcegraph/wildcard'

import styles from './CodeExcerpt.module.scss'

export interface Shape {
    top?: number
    left?: number
    bottom?: number
    right?: number
}

interface Props extends Repo {
    commitID: string
    filePath: string
    highlightRanges: HighlightRange[]
    /** The 0-based (inclusive) line number that this code excerpt starts at */
    startLine: number
    /** The 0-based (exclusive) line number that this code excerpt ends at */
    endLine: number
    className?: string
    /** A function to fetch the range of lines this code excerpt will display. It will be provided
     * the same start and end lines properties that were provided as component props */
    fetchHighlightedFileRangeLines: (startLine: number, endLine: number) => Observable<string[]>
    /** A function to fetch the range of lines this code excerpt will display. It will be provided
     * the same start and end lines properties that were provided as component props */
    fetchPlainTextFileRangeLines?: (startLine: number, endLine: number) => Observable<string[]>
    blobLines: string[]

    visibilityOffset?: Shape
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

const DEFAULT_VISIBILITY_OFFSET: Shape = { bottom: -500 }

/**
 * A code excerpt that displays syntax highlighting and match range highlighting.
 */
export const CodeExcerpt: React.FunctionComponent<Props> = ({
    blobLines,
    fetchHighlightedFileRangeLines,
    startLine,
    endLine,
    highlightRanges,
    visibilityOffset = DEFAULT_VISIBILITY_OFFSET,
    className,
    onCopy,
}) => {
    const [highlightResult, setHighlightResult] = useState<string[] | ErrorLike | undefined>(undefined)

    const [isVisible, setIsVisible] = useState(false)

    const plaintextTable = (
        <table>
            <tbody>
                {blobLines.map((line, i) => (
                    <tr key={line}>
                        <td className="line" data-line={startLine + i + 1} />
                        <td className="code">
                            <span>{line}</span>
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>
    )

    const highlightedTable = isErrorLike(highlightResult)
        ? undefined
        : highlightResult && <table dangerouslySetInnerHTML={{ __html: highlightResult.join('') }} />

    const codeTable = highlightedTable ?? plaintextTable

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

    // Get the syntax highlighted blob lines
    useEffect(() => {
        let subscription: Subscription | undefined
        if (isVisible) {
            const observable = fetchHighlightedFileRangeLines(startLine, endLine)
            subscription = observable.pipe(catchError(error => [asError(error)])).subscribe(setHighlightResult)
        }
        return () => subscription?.unsubscribe()
    }, [startLine, endLine, fetchHighlightedFileRangeLines, isVisible])

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
    }, [highlightRanges, startLine, endLine, tableContainerElement, codeTable])

    return (
        <VisibilitySensor onChange={setIsVisible} partialVisibility={true} offset={visibilityOffset}>
            <Code data-testid="code-excerpt" onCopy={onCopy} className={classNames(styles.codeExcerpt, className)}>
                <div ref={updateTableContainerElementReference}>{codeTable}</div>
            </Code>
        </VisibilitySensor>
    )
}
