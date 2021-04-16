/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-noninteractive-element-interactions: warn */
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
            <DiffBoundary
                diffMode="unified"
                {...hunk}
                contentClassName="diff-hunk__content"
                lineNumbers={lineNumbers}
            />
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
                        className={`diff-hunk__line ${
                            line.kind === DiffHunkLineType.UNCHANGED ? 'diff-hunk__line--both' : ''
                        } ${line.kind === DiffHunkLineType.DELETED ? 'diff-hunk__line--deletion' : ''} ${
                            line.kind === DiffHunkLineType.ADDED ? 'diff-hunk__line--addition' : ''
                        } ${
                            (line.kind !== DiffHunkLineType.ADDED && location.hash === '#' + oldAnchor) ||
                            (line.kind !== DiffHunkLineType.DELETED && location.hash === '#' + newAnchor)
                                ? 'diff-hunk__line--active'
                                : ''
                        }`}
                    >
                        {lineNumbers && (
                            <>
                                {line.kind !== DiffHunkLineType.ADDED ? (
                                    <td
                                        className="diff-hunk__num"
                                        data-line={oldLine - 1}
                                        data-part="base"
                                        id={oldAnchor}
                                    >
                                        {persistLines && (
                                            <Link className="diff-hunk__num--line" to={{ hash: oldAnchor }}>
                                                {oldLine - 1}
                                            </Link>
                                        )}
                                    </td>
                                ) : (
                                    <td className="diff-hunk__num diff-hunk__num--empty" />
                                )}

                                {line.kind !== DiffHunkLineType.DELETED ? (
                                    <td
                                        className="diff-hunk__num"
                                        data-line={newLine - 1}
                                        data-part="head"
                                        id={newAnchor}
                                    >
                                        {persistLines && (
                                            <Link className="diff-hunk__num--line" to={{ hash: newAnchor }}>
                                                {newLine - 1}
                                            </Link>
                                        )}
                                    </td>
                                ) : (
                                    <td className="diff-hunk__num diff-hunk__num--empty" />
                                )}
                            </>
                        )}

                        {/* Needed for decorations */}
                        <td
                            className="diff-hunk__content"
                            /* eslint-disable-next-line react/forbid-dom-props */
                            style={lineStyle}
                            data-diff-marker={diffHunkTypeIndicators[line.kind]}
                        >
                            <div className="d-inline-block" dangerouslySetInnerHTML={{ __html: line.html }} />
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
