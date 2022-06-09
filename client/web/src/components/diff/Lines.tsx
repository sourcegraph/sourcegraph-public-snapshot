import * as React from 'react'

import classNames from 'classnames'
import { DecorationAttachmentRenderOptions, ThemableDecorationStyle } from 'sourcegraph'

import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { decorationAttachmentStyleForTheme } from '@sourcegraph/shared/src/api/extension/api/decorations'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { RouterLink } from '@sourcegraph/wildcard'

import { DiffHunkLineType } from '../../graphql-operations'

import diffHunkStyles from './DiffHunk.module.scss'
import styles from './Lines.module.scss'

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
                hunkContent: styles.lineDeletion,
            }
        case DiffHunkLineType.ADDED:
            return {
                hunkContent: styles.lineAddition,
            }
        default:
            return {
                hunkContent: '',
            }
    }
}

export const EmptyLine: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <>
        <td data-hunk-num={true} className={classNames(diffHunkStyles.numEmpty, diffHunkStyles.num)} />
        <td data-hunk-content-empty={true} className={diffHunkStyles.contentEmpty} />
    </>
)

export const Line: React.FunctionComponent<React.PropsWithChildren<Line>> = ({
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
                    className={classNames(diffHunkStyles.num, hunkStyles.hunkContent, className)}
                    data-line={lineNumber}
                    data-part={dataPart}
                    id={id || anchor}
                    data-hunk-num=" "
                >
                    {persistLines && (
                        <RouterLink className={diffHunkStyles.numLine} to={{ hash: anchor }}>
                            {lineNumber}
                        </RouterLink>
                    )}
                </td>
            )}
            <td
                className={classNames('align-baseline', diffHunkStyles.content, hunkStyles.hunkContent, className)}
                // eslint-disable-next-line react/forbid-dom-props
                style={lineStyle}
                data-diff-marker={diffHunkTypeIndicators[kind]}
            >
                <div className={classNames('d-inline', styles.lineCode)}>
                    <div
                        className={classNames('d-inline-block', styles.lineForceWrap)}
                        dangerouslySetInnerHTML={{ __html: html }}
                        data-diff-marker={diffHunkTypeIndicators[kind]}
                    />
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
