import * as React from 'react'
import { Link } from 'react-router-dom'
import { DecorationAttachmentRenderOptions, ThemableDecorationStyle } from 'sourcegraph'

import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { decorationAttachmentStyleForTheme } from '@sourcegraph/shared/src/api/extension/api/decorations'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { DiffHunkLineType } from '../../graphql-operations'

const diffHunkTypeIndicators: Record<DiffHunkLineType, string> = {
    ADDED: '+',
    UNCHANGED: ' ',
    DELETED: '-',
}

interface Line {
    kind: DiffHunkLineType
    lineNumber?: number
    lineNumbers: boolean
    id?: string
    persistLines: boolean
    anchor: string
    html: string
    lineStyle: ThemableDecorationStyle
    decorations: (TextDocumentDecoration & Record<'after', DecorationAttachmentRenderOptions>)[]
    className: string
}

interface LineType {
    dataPart: string
    hunkContent: string
}

const lineType = (kind: DiffHunkLineType): LineType => {
    switch (kind) {
        case DiffHunkLineType.DELETED:
            return {
                dataPart: 'base',
                hunkContent: 'diff-hunk--split__line--deletion',
            }
        case DiffHunkLineType.UNCHANGED:
            return {
                dataPart: 'base',
                hunkContent: 'diff-hunk--split__line--both',
            }
        case DiffHunkLineType.ADDED:
            return {
                dataPart: 'head',
                hunkContent: 'diff-hunk--split__line--addition',
            }
        default:
            return {
                dataPart: 'base',
                hunkContent: '',
            }
    }
}

export const EmptyLine: React.FunctionComponent = () => (
    <>
        <td className="diff-hunk__num diff-hunk__num--empty" />
        <td className="diff-hunk__content--empty" />
    </>
)

export const Line: React.FunctionComponent<Line> = ({
    persistLines,
    kind,
    lineNumber,
    lineNumbers,
    id,
    anchor,
    html,
    lineStyle,
    decorations,
    className,
}) => {
    const hunkStyles = lineType(kind)

    return (
        <>
            {lineNumbers && (
                <td
                    className={`diff-hunk__num ${hunkStyles.hunkContent}-num ${className}-num`}
                    data-line={lineNumber}
                    data-part={hunkStyles.dataPart}
                    id={id || anchor}
                >
                    {persistLines && (
                        <Link className="diff-hunk__num--line" to={{ hash: anchor }}>
                            {lineNumber}
                        </Link>
                    )}
                </td>
            )}
            <td
                className={`diff-hunk--split__line diff-hunk__content ${hunkStyles.hunkContent} ${className}`}
                // eslint-disable-next-line react/forbid-dom-props
                style={lineStyle}
                data-diff-marker={diffHunkTypeIndicators[kind]}
            >
                <div
                    className="diff-hunk--split__line--code d-inline-block"
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
