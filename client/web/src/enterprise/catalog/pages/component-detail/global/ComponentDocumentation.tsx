import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { CatalogComponentDocumentationFields } from '../../../../../graphql-operations'

interface Props {
    catalogComponent: CatalogComponentDocumentationFields
    className?: string
}

export const ComponentDocumentation: React.FunctionComponent<Props> = ({ catalogComponent: { readme }, className }) => (
    <div className={className}>
        {readme ? (
            <Markdown dangerousInnerHTML={readme.richHTML} />
        ) : (
            <div className="alert alert-warning">No documentation found</div>
        )}
    </div>
)
