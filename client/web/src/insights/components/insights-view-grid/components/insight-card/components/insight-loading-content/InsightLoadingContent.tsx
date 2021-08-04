import { MdiReactIconComponentType } from 'mdi-react'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { InsightDescription } from '../insight-card-description/InsightCardDescription'

interface InsightLoadingContentProps {
    text: string
    subTitle: string
    icon: MdiReactIconComponentType
}

export const InsightLoadingContent: React.FunctionComponent<InsightLoadingContentProps> = props => {
    const { text, subTitle, icon } = props

    return (
        <div className="h-100 w-100 d-flex flex-column">
            <span className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                <LoadingSpinner /> {text}
            </span>
            <InsightDescription className="mt-auto" title={subTitle} icon={icon} />
        </div>
    )
}
