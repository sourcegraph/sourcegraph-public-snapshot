import H from 'history'
import * as React from 'react'
import { ChecksAreaContext } from '../ChecksArea'
import { ChecksManageChecksList } from './ChecksManageChecksList'

interface Props extends ChecksAreaContext {
    history: H.History
    location: H.Location
}

/**
 * The checks management overview page.
 */
export const ChecksManageOverviewPage: React.FunctionComponent<Props> = props => (
    <div className="checks-manage-overview-page">
        <ChecksManageChecksList {...props} />
    </div>
)
