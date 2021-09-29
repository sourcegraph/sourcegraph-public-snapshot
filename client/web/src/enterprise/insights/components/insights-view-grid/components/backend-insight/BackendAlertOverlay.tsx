import classnames from 'classnames'
import ProgressWrench from 'mdi-react/ProgressWrenchIcon'
import React from 'react'

import { AlertOverlay } from '../../../alert-overlay/AlertOverlay'

interface BackendAlertOverLayProps {
    isFetchingHistoricalData?: boolean
    hasNoData: boolean
}
export const BackendAlertOverlay: React.FunctionComponent<BackendAlertOverLayProps> = ({
    isFetchingHistoricalData,
    hasNoData,
}) =>
    isFetchingHistoricalData ? (
        <AlertOverlay
            title="This insight is still being processed"
            description="Datapoints shown may be undercounted."
            icon={<ProgressWrench className={classnames('mb-3')} size={33} />}
        />
    ) : hasNoData ? (
        <AlertOverlay title="No data to display" description="We couldn’t find any matches for this insight." />
    ) : null
