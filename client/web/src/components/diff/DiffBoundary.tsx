import React from 'react'
import { FileDiffHunkFields } from '../../graphql-operations'

interface DiffBoundaryProps extends FileDiffHunkFields {
    lineNumberClassName: string
    contentClassName: string
    lineNumbers: boolean
    diffMode: 'split' | 'unified'
}

interface DiffBoundaryContentProps extends DiffBoundaryProps {
    colspan?: number
}

const DiffBoundaryContent: React.FunctionComponent<DiffBoundaryContentProps> = props => {
    return (
        <>
            {props.lineNumbers && (
                <td className={`diff-boundary__num ${props.lineNumberClassName}`} colSpan={props.colspan} />
            )}
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
}

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
