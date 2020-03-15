import React, { useState, useCallback } from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ChangesetNode, ChangesetNodeProps } from './ChangesetNode'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs, Connection } from '../../../../components/FilteredConnection'
import { Observable, Subject } from 'rxjs'
import { DEFAULT_CHANGESET_LIST_COUNT } from '../presentation'
import { upperFirst } from 'lodash'

interface Props extends ThemeProps {
    queryChangesetsConnection: (
        args: FilteredConnectionQueryArgs
    ) => Observable<Connection<GQL.IExternalChangeset | GQL.IChangesetPlan>>
    history: H.History
    location: H.Location
    campaignUpdates: Subject<void>
    changesetUpdates: Subject<void>
    /** Shows the publish button for ChangesetPlans */
    enablePublishing: ChangesetNodeProps['enablePublishing']
}

/**
 * A list of a campaign's or campaign preview's changesets.
 */
export const CampaignChangesets: React.FunctionComponent<Props> = ({
    queryChangesetsConnection: _queryChangesetsConnection,
    history,
    location,
    isLightTheme,
    changesetUpdates,
    campaignUpdates,
    enablePublishing,
}) => {
    const [state, setState] = useState<GQL.ChangesetState | undefined>()
    const [reviewState, setReviewState] = useState<GQL.ChangesetReviewState | undefined>()
    const [checkState, setCheckState] = useState<GQL.ChangesetCheckState | undefined>()
    const queryChangesetsConnection = useCallback(
        args => _queryChangesetsConnection({ ...args, state, reviewState, checkState }),
        [_queryChangesetsConnection, state, reviewState, checkState]
    )
    return (
        <>
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
                            {upperFirst(state.replace(/_/g, ' ').toLocaleLowerCase())}
                        </option>
                    ))}
                </select>
                <label htmlFor="changeset-review-state-filter">Review state</label>
                <select
                    className="form-control mx-2"
                    value={reviewState}
                    onChange={e =>
                        setReviewState((e.target.value || undefined) as GQL.ChangesetReviewState | undefined)
                    }
                    id="changeset-review-state-filter"
                >
                    <option value="">All</option>
                    {Object.values(GQL.ChangesetReviewState).map(state => (
                        <option value={state} key={state}>
                            {upperFirst(state.replace(/_/g, ' ').toLocaleLowerCase())}
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
                            {upperFirst(state.replace(/_/g, ' ').toLocaleLowerCase())}
                        </option>
                    ))}
                </select>
            </div>
            <div className="list-group">
                <FilteredConnection<GQL.IExternalChangeset | GQL.IChangesetPlan, Omit<ChangesetNodeProps, 'node'>>
                    className="mt-2"
                    updates={changesetUpdates}
                    nodeComponent={ChangesetNode}
                    nodeComponentProps={{ isLightTheme, history, location, campaignUpdates, enablePublishing }}
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
