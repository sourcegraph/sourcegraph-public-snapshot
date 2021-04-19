import React from 'react'

import { FileDiffHunkFields } from '../../graphql-operations'
import { DiffMode } from '../../repo/commit/RepositoryCommitPage'

interface DiffBoundaryProps extends Pick<FileDiffHunkFields, 'oldRange' | 'newRange' | 'section'> {
    contentClassName: string
    lineNumbers: boolean
    diffMode: DiffMode
}

interface DiffBoundaryContentProps extends DiffBoundaryProps {
    colspan?: number
}

const DiffBoundaryContent: React.FunctionComponent<DiffBoundaryContentProps> = props => (
    <>
        {props.lineNumbers && <td className="diff-boundary__num" colSpan={props.colspan} />}
        <td className={`diff-boundary__content ${props.contentClassName}`} data-diff-marker=" ">
            {props.oldRange.lines !== undefined && props.newRange.lines !== undefined && (
                <code className="diff-hunk__line--code diff-hunk__content">
                    @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},
                    {props.newRange.lines} {props.section && `@@ ${props.section}`}
                </code>
            )}
        </td>
    </>
)

export const DiffBoundary: React.FunctionComponent<DiffBoundaryProps> = props => (
    <tr className="diff-boundary">
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
