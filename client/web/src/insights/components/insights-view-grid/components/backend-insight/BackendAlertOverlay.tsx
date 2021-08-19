import classnames from 'classnames'
import ProgressWrench from 'mdi-react/ProgressWrenchIcon'
import React from 'react'

import { AlertOverlay } from '../../../alert-overlay/AlertOverlay'

interface AlertOverLayProps {
    isFetchingHistoricalData?: boolean
    hasNoData: boolean
}
export const BackendAlertOverlay: React.FunctionComponent<AlertOverLayProps> = ({
    isFetchingHistoricalData,
    hasNoData,
}) =>
    isFetchingHistoricalData ? (
        <AlertOverlay title="This insight is still being processed" description="Datapoints shown may be undercounted.">
            <ProgressWrench className={classnames('mb-3')} size={33} />
        </AlertOverlay>
    ) : hasNoData ? (
        <AlertOverlay title="No data to display" description="We couldn’t find any matches for this insight." />
    ) : null
