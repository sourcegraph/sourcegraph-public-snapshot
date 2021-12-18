import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ComponentDocumentationFields } from '../../../../graphql-operations'

interface Props {
    component: ComponentDocumentationFields
    className?: string
}

export const ComponentDocumentation: React.FunctionComponent<Props> = ({ component: { readme }, className }) => (
    <div className={className}>
        {readme ? (
            <Markdown dangerousInnerHTML={readme.richHTML} />
        ) : (
            <div className="alert alert-warning">No documentation found</div>
        )}
    </div>
)
