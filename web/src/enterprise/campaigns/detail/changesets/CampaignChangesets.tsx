import React, { useState, useCallback } from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetNode, ChangesetNodeProps } from './ChangesetNode'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs, Connection } from '../../../../components/FilteredConnection'
import { Observable, Subject } from 'rxjs'
import { DEFAULT_CHANGESET_LIST_COUNT } from '../presentation'
import { upperFirst, lowerCase } from 'lodash'
import { queryChangesetPlans, queryChangesets as _queryChangesets } from '../backend'
import { repeatWhen, delay } from 'rxjs/operators'

interface Props extends ThemeProps {
    campaign: Pick<GQL.ICampaignPlan, '__typename' | 'id'> | Pick<GQL.ICampaign, '__typename' | 'id' | 'closedAt'>
    history: H.History
    location: H.Location
    campaignUpdates: Subject<void>
    changesetUpdates: Subject<void>

    /** For testing only. */
    queryChangesets?: (
        campaignID: GQL.ID,
        args: FilteredConnectionQueryArgs
    ) => Observable<Connection<GQL.IExternalChangeset | GQL.IChangesetPlan>>
}

/**
 * A list of a campaign's or campaign preview's changesets.
 */
export const CampaignChangesets: React.FunctionComponent<Props> = ({
    campaign,
    history,
    location,
    isLightTheme,
    changesetUpdates,
    campaignUpdates,
    queryChangesets = _queryChangesets,
}) => {
    const [state, setState] = useState<GQL.ChangesetState | undefined>()
    const [reviewState, setReviewState] = useState<GQL.ChangesetReviewState | undefined>()
    const [checkState, setCheckState] = useState<GQL.ChangesetCheckState | undefined>()

    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArgs) => {
            const queryObservable: Observable<
                GQL.IChangesetPlanConnection | Connection<GQL.IExternalChangeset | GQL.IChangesetPlan>
            > =
                campaign.__typename === 'CampaignPlan'
                    ? queryChangesetPlans(campaign.id, args)
                    : queryChangesets(campaign.id, { ...args, state, reviewState, checkState })
            return queryObservable.pipe(repeatWhen(obs => obs.pipe(delay(5000))))
        },
        [campaign.id, campaign.__typename, state, reviewState, checkState, queryChangesets]
    )

    const changesetFiltersRow = (
        <div className="form-inline mb-0 mt-2">
            <label htmlFor="changeset-state-filter">State</label>
            <select
                className="form-control mx-2"
                value={state}
                onChange={e => setState((e.target.value || undefined) as GQL.ChangesetState | undefined)}
                id="changeset-state-filter"
            >
                <option value="">All</option>
                {Object.values(GQL.ChangesetState).map(state => (
                    <option value={state} key={state}>
                        {upperFirst(lowerCase(state))}
                    </option>
                ))}
            </select>
            <label htmlFor="changeset-review-state-filter">Review state</label>
            <select
                className="form-control mx-2"
                value={reviewState}
                onChange={e => setReviewState((e.target.value || undefined) as GQL.ChangesetReviewState | undefined)}
                id="changeset-review-state-filter"
            >
                <option value="">All</option>
                {Object.values(GQL.ChangesetReviewState).map(state => (
                    <option value={state} key={state}>
                        {upperFirst(lowerCase(state))}
                    </option>
                ))}
            </select>
            <label htmlFor="changeset-check-state-filter">Check state</label>
            <select
                className="form-control mx-2"
                value={checkState}
                onChange={e => setCheckState((e.target.value || undefined) as GQL.ChangesetCheckState | undefined)}
                id="changeset-check-state-filter"
            >
                <option value="">All</option>
                {Object.values(GQL.ChangesetCheckState).map(state => (
                    <option value={state} key={state}>
                        {upperFirst(lowerCase(state))}
                    </option>
                ))}
            </select>
        </div>
    )

    return (
        <>
            {campaign.__typename === 'Campaign' && changesetFiltersRow}
            <div className="list-group">
                <FilteredConnection<GQL.IExternalChangeset | GQL.IChangesetPlan, Omit<ChangesetNodeProps, 'node'>>
                    className="mt-2"
                    updates={changesetUpdates}
                    nodeComponent={ChangesetNode}
                    nodeComponentProps={{
                        isLightTheme,
                        history,
                        location,
                        campaignUpdates,
                        enablePublishing: campaign.__typename === 'Campaign' && !campaign.closedAt,
                    }}
                    queryConnection={queryChangesetsConnection}
                    hideSearch={true}
                    defaultFirst={DEFAULT_CHANGESET_LIST_COUNT}
                    noun="changeset"
                    pluralNoun="changesets"
                    history={history}
                    location={location}
                    noShowLoaderOnSlowLoad={true}
                    useURLQuery={false}
                />
            </div>
        </>
    )
}
