import React from 'react'
import { FileDiffHunkFields } from '../../graphql-operations'

interface DiffBoundaryProps extends FileDiffHunkFields {
    lineNumberClassName: string
    contentClassName: string
    lineNumbers: boolean
}

export const DiffBoundary: React.FunctionComponent<DiffBoundaryProps> = props => (
    <tr className="diff-boundary">
        {props.lineNumbers && <td className={`diff-boundary__num ${props.lineNumberClassName}`} colSpan={2} />}
        <td className={`diff-boundary__content ${props.contentClassName}`} data-diff-marker=" ">
            {props.oldRange.lines !== undefined && props.newRange.lines !== undefined && (
                <code>
                    @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},
                    {props.newRange.lines} {props.section && `@@ ${props.section}`}
                </code>
            )}
        </td>
    </tr>
)
