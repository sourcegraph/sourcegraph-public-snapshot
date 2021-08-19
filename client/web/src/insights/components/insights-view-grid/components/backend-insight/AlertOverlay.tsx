import classnames from 'classnames'
import ProgressWrench from 'mdi-react/ProgressWrenchIcon'
import React from 'react'

import { BackendAlertOverlay } from './BackendAlertOverlay'

interface AlertOverLayProps {
    isFetchingHistoricalData?: boolean
    hasNoData: boolean
}
export const AlertOverlay: React.FunctionComponent<AlertOverLayProps> = ({ isFetchingHistoricalData, hasNoData }) =>
    isFetchingHistoricalData ? (
        <BackendAlertOverlay
            title="This insight is still being processed"
            description="Datapoints shown may be undercounted."
        >
            <ProgressWrench className={classnames('mb-3')} size={33} />
        </BackendAlertOverlay>
    ) : hasNoData ? (
        <BackendAlertOverlay title="No data to display" description="We couldn’t find any matches for this insight." />
    ) : null
