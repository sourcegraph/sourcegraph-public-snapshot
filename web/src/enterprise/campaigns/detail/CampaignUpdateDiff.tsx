import * as H from 'history'
import * as React from 'react'
import { forkJoin, Observable } from 'rxjs'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ChangesetNode } from './changesets/ChangesetNode'
import { ThemeProps } from '../../../../../shared/src/theme'
import { queryChangesets, queryChangesetPlans } from './backend'
import { useObservable } from '../../../util/useObservable'
import { TabsWithLocalStorageViewStatePersistence } from '../../../../../shared/src/components/Tabs'
import classNames from 'classnames'
import { Connection } from '../../../components/FilteredConnection'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

interface Props extends ThemeProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    campaignPlan: Pick<GQL.ICampaignPlan, 'id'>
    history: H.History
    location: H.Location
    className?: string

    /** Only for testing purposes */
    _queryChangesets?: (
        campaign: GQL.ID,
        { first }: GQL.IChangesetsOnCampaignArguments
    ) => Observable<Connection<GQL.IExternalChangeset | GQL.IChangesetPlan>>
    /** Only for testing purposes */
    _queryChangesetPlans?: (
        campaignPlan: GQL.ID,
        { first }: GQL.IChangesetPlansOnCampaignArguments
    ) => Observable<GQL.IChangesetPlanConnection>
}

/**
 * A list of a campaign's changesets changed over a new plan
 */
export const CampaignUpdateDiff: React.FunctionComponent<Props> = ({
    campaign,
    campaignPlan,
    isLightTheme,
    history,
    location,
    className,
    _queryChangesets = queryChangesets,
    _queryChangesetPlans = queryChangesetPlans,
}) => {
    const queriedChangesets = useObservable(
        React.useMemo(
            () =>
                forkJoin([
                    _queryChangesets(campaign.id, { first: 1000 }),
                    _queryChangesetPlans(campaignPlan.id, { first: 1000 }),
                ]),
            [_queryChangesets, campaign.id, _queryChangesetPlans, campaignPlan.id]
        )
    )
    if (queriedChangesets) {
        const [changesets, changesetPlans] = queriedChangesets
        const changed = changesetPlans.nodes.filter(changesetPlan =>
            changesets.nodes.some(changeset => changeset.repository.id === changesetPlan.repository.id)
        )
        const added = changesetPlans.nodes.filter(
            changesetPlan =>
                !changesets.nodes.some(changeset => changeset.repository.id === changesetPlan.repository.id)
        )
        const deleted = changesets.nodes.filter(
            changeset =>
                changeset.__typename === 'ExternalChangeset' &&
                !changesetPlans.nodes.some(changesetPlan => changesetPlan.repository.id === changeset.repository.id)
        )
        return (
            <TabsWithLocalStorageViewStatePersistence
                storageKey="campaignUpdateDiffTabs"
                className={classNames(className)}
                tabs={[
                    {
                        id: 'added',
                        label: (
                            <span>
                                New changesets <span className="badge badge-secondary badge-pill">{added.length}</span>
                            </span>
                        ),
                    },
                    {
                        id: 'changed',
                        label: (
                            <span>
                                Updated changesets{' '}
                                <span className="badge badge-secondary badge-pill">{changed.length}</span>
                            </span>
                        ),
                    },
                    {
                        id: 'deleted',
                        label: (
                            <span>
                                Closed changesets{' '}
                                <span className="badge badge-secondary badge-pill">{deleted.length}</span>
                            </span>
                        ),
                    },
                ]}
                tabClassName="tab-bar__tab--h5like"
            >
                <div key="added" className="pt-3">
                    {added.map(changeset => (
                        <ChangesetNode
                            enablePublishing={false}
                            history={history}
                            location={location}
                            node={changeset}
                            isLightTheme={isLightTheme}
                            key={changeset.id}
                        />
                    ))}
                    {added.length === 0 && <span className="text-muted">No changesets</span>}
                </div>
                <div key="changed" className="pt-3">
                    {changed.map(changeset => (
                        <ChangesetNode
                            enablePublishing={false}
                            history={history}
                            location={location}
                            node={changeset}
                            isLightTheme={isLightTheme}
                            key={changeset.id}
                        />
                    ))}
                    {changed.length === 0 && <span className="text-muted">No changesets</span>}
                </div>
                <div key="deleted" className="pt-3">
                    {deleted.map(changeset => (
                        <ChangesetNode
                            enablePublishing={false}
                            history={history}
                            location={location}
                            node={changeset}
                            isLightTheme={isLightTheme}
                            key={changeset.id}
                        />
                    ))}
                    {deleted.length === 0 && <span className="text-muted">No changesets</span>}
                </div>
            </TabsWithLocalStorageViewStatePersistence>
        )
    }
    return (
        <p>
            <LoadingSpinner className={classNames('icon-inline', className)} /> Loading diff
        </p>
    )
}
