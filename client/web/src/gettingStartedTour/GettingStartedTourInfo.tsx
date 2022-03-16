import React from 'react'

import classNames from 'classnames'

export const GETTING_STARTED_TOUR_MARKER = 'getting-started-tour-info-marker'

interface GettingStartedTourInfoProps {
    className?: string
}

export const GettingStartedTourInfo: React.FunctionComponent<GettingStartedTourInfoProps> = ({ className }) => (
    <div className={classNames(GETTING_STARTED_TOUR_MARKER, className)} />
)
