import H from 'history'
import * as React from 'react'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { queryChecks } from '../../data'
import { Check } from '../../data'
import { ChecksManageChecksListItem } from './ChecksManageChecksListItem'

interface Props {
    history: H.History
    location: H.Location
}

/**
 * The list of checks in the checks management area.
 */
export const ChecksManageChecksList: React.FunctionComponent<Props> = ({ history, location }) => (
    <div className="checks-manage-checks-list">
        <FilteredConnection<Check>
            listClassName="list-group list-group-flush"
            listComponent="ul"
            noun="check"
            pluralNoun="checks"
            queryConnection={queryChecks}
            nodeComponent={ChecksManageChecksListItem}
            hideSearch={false}
            noSummaryIfAllNodesVisible={true}
            history={history}
            location={location}
        />
    </div>
)
