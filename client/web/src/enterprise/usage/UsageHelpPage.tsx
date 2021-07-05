import React, { useEffect } from 'react'

import { eventLogger } from '../../tracking/eventLogger'

export const UsageHelpPage: React.FunctionComponent = () => {
    useEffect(() => {
        eventLogger.logViewEvent('UsageHelp')
    }, [])

    return (
        <div className="m-3">
            <p className="text-muted">No usage examples</p>
        </div>
    )
}
