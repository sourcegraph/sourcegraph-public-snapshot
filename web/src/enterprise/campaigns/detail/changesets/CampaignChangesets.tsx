import React from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetNode, ChangesetNodeProps } from './ChangesetNode'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { Observable, Subject } from 'rxjs'
import { DEFAULT_CHANGESET_LIST_COUNT } from '../presentation'

interface Props extends ThemeProps {
    queryChangesetsConnection: (
        args: FilteredConnectionQueryArgs
    ) => Observable<GQL.IExternalChangesetConnection | GQL.IChangesetPlanConnection>
    history: H.History
    location: H.Location
    campaignUpdates: Subject<void>
    changesetUpdates: Subject<void>

    className?: string
}

/**
 * A list of a campaign's or campaign preview's changesets.
 */
export const CampaignChangesets: React.FunctionComponent<Props> = ({
    queryChangesetsConnection,
    history,
    location,
    className = '',
    isLightTheme,
    changesetUpdates,
    campaignUpdates
}) => (
    <div className={`list-group ${className}`}>
        <FilteredConnection<GQL.IExternalChangeset | GQL.IChangesetPlan, Omit<ChangesetNodeProps, 'node'>>
            className="mt-2"
            updates={changesetUpdates}
            nodeComponent={ChangesetNode}
            nodeComponentProps={{ isLightTheme, history, location,campaignUpdates }}
            queryConnection={queryChangesetsConnection}
            hideSearch={true}
            defaultFirst={DEFAULT_CHANGESET_LIST_COUNT}
            noun="changeset"
            pluralNoun="changesets"
            history={history}
            location={location}
            useURLQuery={false}
        />
    </div>
)
