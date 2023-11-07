/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-noninteractive-element-interactions: warn */
import * as React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { createLinkUrl, Link } from '@sourcegraph/wildcard'

import { DiffHunkLineType, type FileDiffHunkFields } from '../../graphql-operations'

import { DiffBoundary } from './DiffBoundary'

import styles from './DiffHunk.module.scss'

const diffHunkTypeIndicators: Record<DiffHunkLineType, string> = {
    ADDED: '+',
    UNCHANGED: ' ',
    DELETED: '-',
}
const diffHunkTypeDescriptions: Record<DiffHunkLineType, string> = {
    ADDED: 'added line',
    UNCHANGED: 'unchanged line',
    DELETED: 'deleted line',
}
interface DiffHunkProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string
    hunk: FileDiffHunkFields
    lineNumbers: boolean
    /**
     * Reflect selected line in url
     *
     * @default true
     */
    persistLines?: boolean
}

export const DiffHunk: React.FunctionComponent<React.PropsWithChildren<DiffHunkProps>> = ({
    fileDiffAnchor,
    hunk,
    lineNumbers,
    persistLines = true,
}) => {
    let oldLine = hunk.oldRange.startLine
    let newLine = hunk.newRange.startLine
    const location = useLocation()
    return (
        <>
            <DiffBoundary diffMode="unified" {...hunk} contentClassName={styles.content} lineNumbers={lineNumbers} />
            {hunk.highlight.lines.map((line, index) => {
                if (line.kind !== DiffHunkLineType.ADDED) {
                    oldLine++
                }
                if (line.kind !== DiffHunkLineType.DELETED) {
                    newLine++
                }
                const oldAnchor = `${fileDiffAnchor}L${oldLine - 1}`
                const newAnchor = `${fileDiffAnchor}R${newLine - 1}`
                return (
                    /*
                        a11y-ignore
                        Rule: "color-contrast" (Elements must have sufficient color contrast) for all changes in this file
                        GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                    */
                    <tr
                        key={index}
                        data-hunk-line-kind={line.kind}
                        className={classNames(
                            'a11y-ignore',
                            line.kind === DiffHunkLineType.UNCHANGED && styles.lineBoth,
                            line.kind === DiffHunkLineType.DELETED && styles.lineDeletion,
                            line.kind === DiffHunkLineType.ADDED && styles.lineAddition,
                            ((line.kind !== DiffHunkLineType.ADDED && location.hash === '#' + oldAnchor) ||
                                (line.kind !== DiffHunkLineType.DELETED && location.hash === '#' + newAnchor)) &&
                            styles.lineActive
                        )}
                    >
                        {lineNumbers && (
                            <>
                                {line.kind !== DiffHunkLineType.ADDED ? (
                                    <td
                                        className={styles.num}
                                        data-line={oldLine - 1}
                                        data-part="base"
                                        id={oldAnchor}
                                        data-hunk-num=" "
                                    >
                                        {persistLines && (
                                            <Link className={styles.numLine} to={createLinkUrl({ hash: oldAnchor })}>
                                                {oldLine - 1}
                                            </Link>
                                        )}
                                    </td>
                                ) : (
                                    <td className={classNames(styles.num, styles.numEmpty)} />
                                )}

                                {line.kind !== DiffHunkLineType.DELETED ? (
                                    <td
                                        className={styles.num}
                                        data-line={newLine - 1}
                                        data-part="head"
                                        id={newAnchor}
                                        data-hunk-num=" "
                                    >
                                        {persistLines && (
                                            <Link className={styles.numLine} to={createLinkUrl({ hash: newAnchor })}>
                                                {newLine - 1}
                                            </Link>
                                        )}
                                    </td>
                                ) : (
                                    <td className={classNames(styles.num, styles.numEmpty)} />
                                )}
                            </>
                        )}
                        <td className={styles.content}>
                            <VisuallyHidden>{diffHunkTypeDescriptions[line.kind]}</VisuallyHidden>
                            <div
                                className="d-inline-block"
                                data-diff-marker={diffHunkTypeIndicators[line.kind]}
                                dangerouslySetInnerHTML={{ __html: line.html }}
                            />
                        </td>
                    </tr>
                )
            })}
        </>
    )
}
