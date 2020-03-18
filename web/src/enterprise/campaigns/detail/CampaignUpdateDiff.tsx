import * as H from 'history'
import * as React from 'react'
import { forkJoin, Observable } from 'rxjs'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ChangesetNode } from './changesets/ChangesetNode'
import { ThemeProps } from '../../../../../shared/src/theme'
import { queryChangesets, queryChangesetPlans } from './backend'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { TabsWithLocalStorageViewStatePersistence } from '../../../../../shared/src/components/Tabs'
import classNames from 'classnames'
import { Connection } from '../../../components/FilteredConnection'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { pluralize } from '../../../../../shared/src/util/strings'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

interface Props extends ThemeProps {
    campaign: Pick<GQL.ICampaign, 'id' | 'publishedAt'> & {
        changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'>
    } & {
        changesetPlans: Pick<GQL.ICampaign['changesetPlans'], 'totalCount'>
    }
    campaignPlan: Pick<GQL.ICampaignPlan, 'id'> & {
        changesetPlans: Pick<GQL.ICampaignPlan['changesetPlans'], 'totalCount'>
    }
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

export type ChangesetArray = (GQL.IExternalChangeset | GQL.IChangesetPlan)[]

export interface CampaignDiff {
    added: ChangesetArray
    changed: ChangesetArray
    /**
     * Unmodified are all changesets, that need to be updated via gitserver.
     * Changing the campaign description will technically update them,
     * but they will still show up as "unmodified" to reduce confusion
     */
    unmodified: ChangesetArray
    deleted: ChangesetArray
}

export function calculateChangesetDiff(changesets: ChangesetArray, changesetPlans: GQL.IChangesetPlan[]): CampaignDiff {
    const added: ChangesetArray = []
    const changed: ChangesetArray = []
    const unmodified: ChangesetArray = []
    const deleted: ChangesetArray = []

    const changesetsByRepoId = new Map<string, GQL.IExternalChangeset | GQL.IChangesetPlan>()
    for (const changeset of changesets) {
        changesetsByRepoId.set(changeset.repository.id, changeset)
    }
    for (const changesetPlan of changesetPlans) {
        const key = changesetPlan.repository.id
        const existingChangeset = changesetsByRepoId.get(key)
        // if no matching changeset exists yet, it is a new changeset to the campaign
        if (!existingChangeset) {
            added.push(changesetPlan)
            continue
        }
        changesetsByRepoId.delete(key)
        // if the matching changeset has not been published yet, or the existing changeset is still open, it will be updated
        if (
            existingChangeset.__typename === 'ChangesetPlan' ||
            ![GQL.ChangesetState.MERGED, GQL.ChangesetState.CLOSED].includes(existingChangeset.state)
        ) {
            changed.push(changesetPlan)
            continue
        }
        unmodified.push(existingChangeset)
    }
    for (const changeset of changesetsByRepoId.values()) {
        if (changeset.__typename === 'ChangesetPlan') {
            // don't mention any preexisting changesetplans that don't apply anymore
            continue
        }
        if ([GQL.ChangesetState.MERGED, GQL.ChangesetState.CLOSED].includes(changeset.state)) {
            unmodified.push(changeset)
        } else {
            deleted.push(changeset)
        }
    }

    return {
        added,
        changed,
        unmodified,
        deleted,
    }
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
        const { added, changed, unmodified, deleted } = calculateChangesetDiff(changesets.nodes, changesetPlans.nodes)

        const newDraftCount = !campaign.publishedAt
            ? changed.length - (campaign.changesets.totalCount - deleted.length) + added.length
            : 0
        return (
            <div className={className}>
                <h3 className="mt-4 mb-2">Preview of changes</h3>
                <p>
                    Campaign currently has {campaign.changesets.totalCount + campaign.changesetPlans.totalCount}{' '}
                    {pluralize('changeset', campaign.changesets.totalCount + campaign.changesetPlans.totalCount)} (
                    {campaign.changesets.totalCount} published, {campaign.changesetPlans.totalCount}{' '}
                    {pluralize('draft', campaign.changesetPlans.totalCount)}), after update it will have{' '}
                    {campaignPlan.changesetPlans.totalCount}{' '}
                    {pluralize('changeset', campaignPlan.changesetPlans.totalCount)} (
                    {campaign.publishedAt
                        ? unmodified.length + changed.length - deleted.length + added.length
                        : campaign.changesets.totalCount - deleted.length}{' '}
                    published, {newDraftCount} {pluralize('draft', newDraftCount)}):
                </p>
                <TabsWithLocalStorageViewStatePersistence
                    storageKey="campaignUpdateDiffTabs"
                    tabs={[
                        {
                            id: 'added',
                            label: (
                                <span>
                                    To be created{' '}
                                    <span className="badge badge-secondary badge-pill">{added.length}</span>
                                </span>
                            ),
                        },
                        {
                            id: 'changed',
                            label: (
                                <span>
                                    To be updated{' '}
                                    <span className="badge badge-secondary badge-pill">{changed.length}</span>
                                </span>
                            ),
                        },
                        {
                            id: 'unmodified',
                            label: (
                                <span>
                                    Unmodified{' '}
                                    <span className="badge badge-secondary badge-pill">{unmodified.length}</span>
                                </span>
                            ),
                        },
                        {
                            id: 'deleted',
                            label: (
                                <span>
                                    To be closed{' '}
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
                    <div key="unmodified" className="pt-3">
                        {unmodified.map(changeset => (
                            <ChangesetNode
                                enablePublishing={false}
                                history={history}
                                location={location}
                                node={changeset}
                                isLightTheme={isLightTheme}
                                key={changeset.id}
                            />
                        ))}
                        {unmodified.length === 0 && <span className="text-muted">No changesets</span>}
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
                <div className="alert alert-info mt-2">
                    <AlertCircleIcon className="icon-inline" /> You are updating an existing campaign. By clicking
                    'Update', all above changesets that are not 'unmodified' will be updated on the codehost.
                </div>
            </div>
        )
    }
    return (
        <div>
            <LoadingSpinner className={classNames('icon-inline', className)} /> Loading diff
        </div>
    )
}
