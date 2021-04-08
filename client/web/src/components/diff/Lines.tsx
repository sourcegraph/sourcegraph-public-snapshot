import * as React from 'react'
import { decorationAttachmentStyleForTheme } from '../../../../shared/src/api/extension/api/decorations'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { Link } from 'react-router-dom'
import { DecorationAttachmentRenderOptions, TextDocumentDecoration, ThemableDecorationStyle } from 'sourcegraph'
import { DiffHunkLineType } from '../../graphql-operations'

const diffHunkTypeIndicators: Record<DiffHunkLineType, string> = {
    ADDED: '+',
    UNCHANGED: ' ',
    DELETED: '-',
}

interface Line {
    kind: DiffHunkLineType
    lineNumber?: number
    anchor: string
    html: string
    lineStyle: ThemableDecorationStyle
    decorations: (TextDocumentDecoration & Record<'after', DecorationAttachmentRenderOptions>)[]
    className: string
}

export const EmptyLine: React.FunctionComponent = () => (
    <>
        <td className="diff-hunk__num diff-hunk__num--empty" />
        <td className="diff-hunk__content--empty" />
    </>
)

export const Line: React.FunctionComponent<Line> = ({
    kind,
    lineNumber,
    anchor,
    html,
    lineStyle,
    decorations,
    className,
}) => {
    const lineType = (
        type: DiffHunkLineType
    ): {
        dataPart: string
        hunkContent: string
        hash: string
    } => {
        switch (type) {
            case DiffHunkLineType.DELETED:
                return {
                    dataPart: 'base',
                    hunkContent: 'diff-hunk__line--deletion',
                    hash: anchor,
                }
            case DiffHunkLineType.UNCHANGED:
                return {
                    dataPart: 'base',
                    hunkContent: 'diff-hunk__line--both',
                    hash: anchor,
                }
            case DiffHunkLineType.ADDED:
                return {
                    dataPart: 'head',
                    hunkContent: 'diff-hunk__line--addition',
                    hash: anchor,
                }
            default:
                return {
                    dataPart: 'base',
                    hunkContent: '',
                    hash: '',
                }
        }
    }

    const hunkStyles = lineType(kind)

    return (
        <>
            <td
                className={`diff-hunk__num ${className}`}
                data-line={lineNumber}
                data-part={hunkStyles.dataPart}
                id={hunkStyles.hash}
            >
                <Link className="diff-hunk__num--data-line" to={{ hash: hunkStyles.hash }}>
                    {lineNumber}
                </Link>
            </td>
            <td
                className={`diff-hunk__line diff-hunk__content ${hunkStyles.hunkContent} ${className}`}
                // eslint-disable-next-line react/forbid-dom-props
                style={lineStyle}
                data-diff-marker={diffHunkTypeIndicators[kind]}
            >
                <span
                    className="diff-hunk__line--code"
                    dangerouslySetInnerHTML={{ __html: html }}
                    data-diff-marker={diffHunkTypeIndicators[kind]}
                />
                {decorations.map((decoration, index) => {
                    const style = decorationAttachmentStyleForTheme(decoration.after, true)
                    return (
                        <React.Fragment key={index}>
                            {' '}
                            <LinkOrSpan
                                className="d-inline-block"
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
        </>
    )
}
