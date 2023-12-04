import React from 'react'

import classNames from 'classnames'

import { Code } from '@sourcegraph/wildcard'

import type { FileDiffHunkFields } from '../../graphql-operations'
import type { DiffMode } from '../../repo/commit/RepositoryCommitPage'

import styles from './DiffBoundary.module.scss'

interface DiffBoundaryProps extends Pick<FileDiffHunkFields, 'oldRange' | 'newRange' | 'section'> {
    contentClassName: string
    lineNumbers: boolean
    diffMode: DiffMode
}

interface DiffBoundaryContentProps extends DiffBoundaryProps {
    colspan?: number
}

const DiffBoundaryContent: React.FunctionComponent<React.PropsWithChildren<DiffBoundaryContentProps>> = props => (
    <>
        {props.lineNumbers && <td className={styles.num} data-diff-boundary-num={true} colSpan={props.colspan} />}
        <td
            className={classNames(styles.content, props.contentClassName)}
            data-diff-boundary-content={true}
            data-diff-marker=" "
        >
            {props.oldRange.lines !== undefined && props.newRange.lines !== undefined && (
                <Code>
                    @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},
                    {props.newRange.lines} {props.section && `@@ ${props.section}`}
                </Code>
            )}
        </td>
    </>
)

export const DiffBoundary: React.FunctionComponent<React.PropsWithChildren<DiffBoundaryProps>> = props => (
    <tr>
        {props.diffMode === 'split' ? (
            <>
                <DiffBoundaryContent {...props} />
                <DiffBoundaryContent {...props} />
            </>
        ) : (
            <DiffBoundaryContent {...props} colspan={2} />
        )}
    </tr>
)
