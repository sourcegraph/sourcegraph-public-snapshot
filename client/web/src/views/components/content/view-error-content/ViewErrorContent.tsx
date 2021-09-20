import { MdiReactIconComponentType } from 'mdi-react'
import React from 'react'

import { ErrorLike } from '@sourcegraph/codeintellify/lib/errors'

import { ErrorAlert } from '../../../../components/alerts'
import { ViewCardDescription } from '../../card/view-card-description/ViewCardDescription'

interface ViewErrorContentProps {
    title: string
    error: ErrorLike
    icon: MdiReactIconComponentType
}

export const ViewErrorContent: React.FunctionComponent<ViewErrorContentProps> = props => {
    const { error, title, icon, children } = props

    return (
        <div className="h-100 w-100 d-flex flex-column">
            {children || <ErrorAlert data-testid={`${title} view error`} className="m-0" error={error} />}
            <ViewCardDescription className="mt-auto" title={title} icon={icon} />
        </div>
    )
}
