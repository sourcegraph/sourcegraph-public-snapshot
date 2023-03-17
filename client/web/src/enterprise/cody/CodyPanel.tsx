import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

export const CodyPanel: React.FunctionComponent<
    {
        repoID: string
        revision?: string
        filePath: string
    } & TelemetryProps
> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.log('CodyPanelOpened')
    }, [telemetryService])

    return <div>CodyPanel</div>
}
