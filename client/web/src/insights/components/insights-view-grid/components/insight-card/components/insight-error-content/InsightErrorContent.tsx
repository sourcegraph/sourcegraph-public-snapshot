import { MdiReactIconComponentType } from 'mdi-react'
import React from 'react'

import { ErrorLike } from '@sourcegraph/codeintellify/lib/errors'

import { ErrorAlert } from '../../../../../../../components/alerts'
import { InsightDescription } from '../insight-card-description/InsightCardDescription'

interface InsightErrorContentProps {
    title: string
    error: ErrorLike
    icon: MdiReactIconComponentType
}

export const InsightErrorContent: React.FunctionComponent<InsightErrorContentProps> = props => {
    const { error, title, icon, children } = props

    return (
        <div className="h-100 w-100 d-flex flex-column">
            {children || <ErrorAlert data-testid={`${title} insight error`} className="m-0" error={error} />}
            <InsightDescription className="mt-auto" title={title} icon={icon} />
        </div>
    )
}
