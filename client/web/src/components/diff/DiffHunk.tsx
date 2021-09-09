/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-noninteractive-element-interactions: warn */
import classNames from 'classnames'
import * as React from 'react'
import { useLocation } from 'react-router'
import { Link } from 'react-router-dom'

import {
    decorationAttachmentStyleForTheme,
    DecorationMapByLine,
    decorationStyleForTheme,
} from '@sourcegraph/shared/src/api/extension/api/decorations'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined, property } from '@sourcegraph/shared/src/util/types'

import { DiffHunkLineType, FileDiffHunkFields } from '../../graphql-operations'

import { DiffBoundary } from './DiffBoundary'
import styles from './DiffHunk.module.scss'

const diffHunkTypeIndicators: Record<DiffHunkLineType, string> = {
    ADDED: '+',
    UNCHANGED: ' ',
    DELETED: '-',
}
interface DiffHunkProps extends ThemeProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string
    hunk: FileDiffHunkFields
    lineNumbers: boolean
    decorations: Record<'head' | 'base', DecorationMapByLine>
    /**
     * Reflect selected line in url
     *
     * @default true
     */
    persistLines?: boolean
}

export const DiffHunk: React.FunctionComponent<DiffHunkProps> = ({
    fileDiffAnchor,
    decorations,
    hunk,
    lineNumbers,
    persistLines = true,
    isLightTheme,
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
                const decorationsForLine = [
                    // If the line was deleted, look for decorations in the base revision
                    ...((line.kind === DiffHunkLineType.DELETED && decorations.base.get(oldLine - 1)) || []),
                    // If the line wasn't deleted, look for decorations in the head revision
                    ...((line.kind !== DiffHunkLineType.DELETED && decorations.head.get(newLine - 1)) || []),
                ]
                const lineStyle = decorationsForLine
                    .filter(decoration => decoration.isWholeLine)
                    .map(decoration => decorationStyleForTheme(decoration, isLightTheme))
                    .reduce((style, decoration) => ({ ...style, ...decoration }), {})
                return (
                    <tr
                        key={index}
                        data-hunk-line-kind={line.kind}
                        className={classNames(
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
                                            <Link className={styles.numLine} to={{ hash: oldAnchor }}>
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
                                            <Link className={styles.numLine} to={{ hash: newAnchor }}>
                                                {newLine - 1}
                                            </Link>
                                        )}
                                    </td>
                                ) : (
                                    <td className={classNames(styles.num, styles.numEmpty)} />
                                )}
                            </>
                        )}

                        {/* Needed for decorations */}
                        <td
                            className={styles.content}
                            /* eslint-disable-next-line react/forbid-dom-props */
                            style={lineStyle}
                            data-diff-marker={diffHunkTypeIndicators[line.kind]}
                        >
                            <div className="d-inline" dangerouslySetInnerHTML={{ __html: line.html }} />
                            {decorationsForLine.filter(property('after', isDefined)).map((decoration, index) => {
                                const style = decorationAttachmentStyleForTheme(decoration.after, isLightTheme)
                                return (
                                    <React.Fragment key={index}>
                                        {' '}
                                        <LinkOrSpan
                                            to={decoration.after.linkURL}
                                            data-tooltip={decoration.after.hoverMessage}
                                            style={style}
                                        >
                                            {decoration.after.contentText}
                                        </LinkOrSpan>
                                    </React.Fragment>
                                )
                            })}
                        </td>
                    </tr>
                )
            })}
        </>
    )
}
