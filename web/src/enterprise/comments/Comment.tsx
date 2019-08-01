import React from 'react'
import { Markdown } from '../../../../shared/src/components/Markdown'
import * as GQL from '../../../../shared/src/graphql/schema'

interface Props {
    comment: GQL.IComment

    className?: string
}

/**
 * A comment with a rendered body and an edit mode.
 */
export const Comment: React.FunctionComponent<Props> = ({ comment, className = '' }) => {
    const a = 1
    return (
        <div className={`comment ${className}`}>
            <Markdown dangerousInnerHTML={comment.bodyHTML} />
        </div>
    )
}
