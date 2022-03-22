import React from 'react'

import classNames from 'classnames'

export const GETTING_STARTED_TOUR_MARKER = 'getting-started-tour-info-marker'

interface GettingStartedTourInfoProps {
    className?: string
    isSourcegraphDotCom: boolean
}

export const GettingStartedTourInfo: React.FunctionComponent<GettingStartedTourInfoProps> = ({
    className,
    isSourcegraphDotCom,
}) => {
    if (!isSourcegraphDotCom) {
        return null
    }
    return <div className={classNames(GETTING_STARTED_TOUR_MARKER, className)} />
}
