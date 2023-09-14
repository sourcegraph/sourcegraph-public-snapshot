import React, { useCallback, useMemo } from 'react'

import classNames from 'classnames'
import { ReplaySubject } from 'rxjs'

import type { FileDiffFields } from '../../graphql-operations'
import type { DiffMode } from '../../repo/commit/RepositoryCommitPage'

import { DiffHunk } from './DiffHunk'
import { DiffSplitHunk } from './DiffSplitHunk'

import styles from './FileDiffHunks.module.scss'

export interface FileHunksProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string

    /** The file's hunks. */
    hunks: FileDiffFields['hunks']

    /** Whether to show line numbers. */
    lineNumbers: boolean

    className: string
    /** Reflect selected line in url */
    persistLines?: boolean
    diffMode: DiffMode
}

/** Displays hunks in a unified file diff. */
export const FileDiffHunks: React.FunctionComponent<React.PropsWithChildren<FileHunksProps>> = ({
    className,
    fileDiffAnchor,
    hunks,
    lineNumbers,
    persistLines,
    diffMode,
}) => {
    /** Emits whenever the ref callback for the code element is called */
    const codeElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextCodeElement = useCallback(
        (element: HTMLElement | null): void => codeElements.next(element),
        [codeElements]
    )

    /** Emits whenever the ref callback for the blob element is called */
    const blobElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextBlobElement = useCallback(
        (element: HTMLElement | null): void => blobElements.next(element),
        [blobElements]
    )

    const isSplitMode = diffMode === 'split'

    return (
        <div className={styles.body}>
            <div className={classNames(styles.fileDiffHunks, className)} ref={nextBlobElement}>
                {hunks.length === 0 ? (
                    <div className="text-muted m-2">No changes</div>
                ) : (
                    <div className={styles.container} ref={nextCodeElement}>
                        <table
                            className={classNames(styles.table, {
                                [styles.tableSplit]: isSplitMode,
                            })}
                        >
                            {lineNumbers && (
                                <colgroup>
                                    {isSplitMode ? (
                                        <>
                                            <col width="40" />
                                            <col />
                                            <col width="40" />
                                            <col />
                                        </>
                                    ) : (
                                        <>
                                            <col width="40" />
                                            <col width="40" />
                                            <col />
                                        </>
                                    )}
                                </colgroup>
                            )}
                            <tbody>
                                {hunks.map(hunk =>
                                    isSplitMode ? (
                                        <DiffSplitHunk
                                            fileDiffAnchor={fileDiffAnchor}
                                            lineNumbers={lineNumbers}
                                            persistLines={persistLines}
                                            key={hunk.oldRange.startLine}
                                            hunk={hunk}
                                        />
                                    ) : (
                                        <DiffHunk
                                            fileDiffAnchor={fileDiffAnchor}
                                            lineNumbers={lineNumbers}
                                            persistLines={persistLines}
                                            key={hunk.oldRange.startLine}
                                            hunk={hunk}
                                        />
                                    )
                                )}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    )
}
