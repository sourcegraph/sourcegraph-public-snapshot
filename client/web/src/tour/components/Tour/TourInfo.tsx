import React from 'react'

import classNames from 'classnames'

export const GETTING_STARTED_TOUR_MARKER = 'getting-started-tour-info-marker'

export interface TourInfoProps {
    className?: string
    isSourcegraphDotCom: boolean
}

export const TourInfo: React.FunctionComponent<TourInfoProps> = ({ className, isSourcegraphDotCom }) => {
    if (!isSourcegraphDotCom) {
        return null
    }
    return <div className={classNames(GETTING_STARTED_TOUR_MARKER, className)} />
}
