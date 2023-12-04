import * as React from 'react'

import classNames from 'classnames'

import { createLinkUrl, RouterLink } from '@sourcegraph/wildcard'

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
    className: string
    dataPart: 'head' | 'base'
}

interface LineType {
    hunkContent: string
}

const lineType = (kind: DiffHunkLineType): LineType => {
    switch (kind) {
        case DiffHunkLineType.DELETED: {
            return {
                hunkContent: styles.lineDeletion,
            }
        }
        case DiffHunkLineType.ADDED: {
            return {
                hunkContent: styles.lineAddition,
            }
        }
        default: {
            return {
                hunkContent: '',
            }
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
    className,
    dataPart,
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
                        <RouterLink className={diffHunkStyles.numLine} to={createLinkUrl({ hash: anchor })}>
                            {lineNumber}
                        </RouterLink>
                    )}
                </td>
            )}
            <td
                className={classNames('align-baseline', diffHunkStyles.content, hunkStyles.hunkContent, className)}
                data-diff-marker={diffHunkTypeIndicators[kind]}
            >
                <div className={classNames('d-inline', styles.lineCode)}>
                    <div
                        className={classNames('d-inline-block', styles.lineForceWrap)}
                        dangerouslySetInnerHTML={{ __html: html }}
                        data-diff-marker={diffHunkTypeIndicators[kind]}
                    />
                </div>
            </td>
        </>
    )
}
