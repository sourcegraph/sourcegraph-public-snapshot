import React, { useEffect } from 'react'

import { eventLogger } from '../../tracking/eventLogger'

export const UsageHelpPage: React.FunctionComponent = () => {
    useEffect(() => {
        eventLogger.logViewEvent('UsageHelp')
    }, [])

    return (
        <div>
            <h1>Usage</h1>
            <p>Welcome!</p>
        </div>
    )
}
