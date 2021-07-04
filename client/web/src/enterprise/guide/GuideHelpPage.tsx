import React, { useEffect } from 'react'

import { eventLogger } from '../../tracking/eventLogger'

export const GuideHelpPage: React.FunctionComponent = () => {
    useEffect(() => {
        eventLogger.logViewEvent('GuideHelp')
    }, [])

    return (
        <div>
            <h1>Sourcegraph Guide</h1>
            <p>Welcome!</p>
        </div>
    )
}
