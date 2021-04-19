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
    dataPart: 'head' | 'base'
    isLightTheme: boolean
}

interface LineType {
    hunkContent: string
}

const lineType = (kind: DiffHunkLineType): LineType => {
    switch (kind) {
        case DiffHunkLineType.DELETED:
            return {
                hunkContent: 'diff-hunk--split__line--deletion',
            }
        case DiffHunkLineType.UNCHANGED:
            return {
                hunkContent: 'diff-hunk--split__line--both',
            }
        case DiffHunkLineType.ADDED:
            return {
                hunkContent: 'diff-hunk--split__line--addition',
            }
        default:
            return {
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
    dataPart,
    isLightTheme,
}) => {
    const hunkStyles = lineType(kind)

    return (
        <>
            {lineNumbers && (
                <td
                    className={`diff-hunk__num ${hunkStyles.hunkContent}-num ${className}-num`}
                    data-line={lineNumber}
                    data-part={dataPart}
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
                <div className="diff-hunk--split__line--code d-inline-block">
                    <div dangerouslySetInnerHTML={{ __html: html }} data-diff-marker={diffHunkTypeIndicators[kind]} />
                    {decorations.map((decoration, index) => {
                        const style = decorationAttachmentStyleForTheme(decoration.after, isLightTheme)
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
                </div>
            </td>
        </>
    )
}
