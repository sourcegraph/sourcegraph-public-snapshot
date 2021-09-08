import { MdiReactIconComponentType } from 'mdi-react'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { ViewCardDescription } from '../../card/view-card-description/ViewCardDescription'

interface ViewLoadingContentProps {
    text: string
    subTitle: string
    icon: MdiReactIconComponentType
}

export const ViewLoadingContent: React.FunctionComponent<ViewLoadingContentProps> = props => {
    const { text, subTitle, icon } = props

    return (
        <div className="h-100 w-100 d-flex flex-column">
            <span className="flex-grow-1 d-flex flex-column align-items-center justify-content-center">
                <LoadingSpinner /> {text}
            </span>
            <ViewCardDescription className="mt-auto" title={subTitle} icon={icon} />
        </div>
    )
}
