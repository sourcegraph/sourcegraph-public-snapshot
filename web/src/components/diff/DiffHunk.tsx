import * as H from 'history'
import * as React from 'react'
import {
    decorationAttachmentStyleForTheme,
    DecorationMapByLine,
    decorationStyleForTheme,
} from '../../../../shared/src/api/client/services/decoration'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import * as GQL from '../../../../shared/src/graphql/schema'
import { property, isDefined } from '../../../../shared/src/util/types'
import { ThemeProps } from '../../../../shared/src/theme'

const DiffBoundary: React.FunctionComponent<{
    /** The "lines" property is set for end boundaries (only for start boundaries and between hunks). */
    oldRange: {
        startLine: number
        lines?: number
    }
    newRange: {
        startLine: number
        lines?: number
    }
    section: string | null
    lineNumberClassName: string
    contentClassName: string
    lineNumbers: boolean
}> = props => (
    <tr className="diff-boundary">
        {props.lineNumbers && <td className={`diff-boundary__num ${props.lineNumberClassName}`} colSpan={2} />}
        <td className={`diff-boundary__content ${props.contentClassName}`}>
            {props.oldRange.lines !== undefined && props.newRange.lines !== undefined && (
                <code>
                    @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},
                    {props.newRange.lines} {props.section && `@@ ${props.section}`}
                </code>
            )}
        </td>
    </tr>
)
export const DiffHunk: React.FunctionComponent<
    {
        /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
        fileDiffAnchor: string
        hunk: GQL.IFileDiffHunk
        lineNumbers: boolean
        decorations: Record<'head' | 'base', DecorationMapByLine>
        location: H.Location
        history: H.History
        /**
         * Reflect selected line in url
         *
         * @default true
         */
        persistLines?: boolean
    } & ThemeProps
> = ({ fileDiffAnchor, decorations, hunk, lineNumbers, location, history, persistLines = true, isLightTheme }) => {
    let oldLine = hunk.oldRange.startLine
    let newLine = hunk.newRange.startLine
    return (
        <>
            <DiffBoundary
                {...hunk}
                lineNumberClassName="diff-hunk__num--both"
                contentClassName="diff-hunk__content"
                lineNumbers={lineNumbers}
            />
            {hunk.highlight.lines.map((line, i) => {
                if (line.kind !== GQL.DiffHunkLineType.ADDED) {
                    oldLine++
                }
                if (line.kind !== GQL.DiffHunkLineType.DELETED) {
                    newLine++
                }
                const oldAnchor = `${fileDiffAnchor}L${oldLine - 1}`
                const newAnchor = `${fileDiffAnchor}R${newLine - 1}`
                const decorationsForLine = [
                    // If the line was deleted, look for decorations in the base rev
                    ...((line.kind === GQL.DiffHunkLineType.DELETED && decorations.base.get(oldLine - 1)) || []),
                    // If the line wasn't deleted, look for decorations in the head rev
                    ...((line.kind !== GQL.DiffHunkLineType.DELETED && decorations.head.get(newLine - 1)) || []),
                ]
                const lineStyle = decorationsForLine
                    .filter(decoration => decoration.isWholeLine)
                    .map(decoration => decorationStyleForTheme(decoration, isLightTheme))
                    .reduce((style, decoration) => ({ ...style, ...decoration }), {})
                return (
                    <tr
                        key={i}
                        className={`diff-hunk__line ${
                            line.kind === GQL.DiffHunkLineType.UNCHANGED ? 'diff-hunk__line--both' : ''
                        } ${line.kind === GQL.DiffHunkLineType.DELETED ? 'diff-hunk__line--deletion' : ''} ${
                            line.kind === GQL.DiffHunkLineType.ADDED ? 'diff-hunk__line--addition' : ''
                        } ${
                            (line.kind !== GQL.DiffHunkLineType.ADDED && location.hash === '#' + oldAnchor) ||
                            (line.kind !== GQL.DiffHunkLineType.DELETED && location.hash === '#' + newAnchor)
                                ? 'diff-hunk__line--active'
                                : ''
                        }`}
                    >
                        {lineNumbers && (
                            <>
                                {line.kind !== GQL.DiffHunkLineType.ADDED ? (
                                    <td
                                        className="diff-hunk__num"
                                        data-line={oldLine - 1}
                                        data-part="base"
                                        id={oldAnchor}
                                        onClick={() => persistLines && history.push({ hash: oldAnchor })}
                                    />
                                ) : (
                                    <td className="diff-hunk__num diff-hunk__num--empty" />
                                )}

                                {line.kind !== GQL.DiffHunkLineType.DELETED ? (
                                    <td
                                        className="diff-hunk__num"
                                        data-line={newLine - 1}
                                        data-part="head"
                                        id={newAnchor}
                                        onClick={() => persistLines && history.push({ hash: newAnchor })}
                                    />
                                ) : (
                                    <td className="diff-hunk__num diff-hunk__num--empty" />
                                )}
                            </>
                        )}

                        {/* Needed for decorations */}
                        {/* eslint-disable-next-line react/forbid-dom-props */}
                        <td className="diff-hunk__content" style={lineStyle}>
                            <div className="d-inline-block" dangerouslySetInnerHTML={{ __html: line.html }} />
                            {decorationsForLine.filter(property('after', isDefined)).map((decoration, i) => {
                                const style = decorationAttachmentStyleForTheme(decoration.after, isLightTheme)
                                return (
                                    <React.Fragment key={i}>
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
