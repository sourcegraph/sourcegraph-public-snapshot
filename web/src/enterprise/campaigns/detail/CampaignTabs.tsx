import React, { useMemo, useCallback } from 'react'
import H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../shared/src/theme'
import { TabsWithLocalStorageViewStatePersistence } from '../../../../../shared/src/components/Tabs'
import { CampaignDiffs } from './diffs/CampaignDiffs'
import { CampaignChangesets } from './changesets/CampaignChangesets'
import { queryChangesets, queryChangesetPlans } from './backend'
import { FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { Subject } from 'rxjs'

interface Props extends ThemeProps {
    campaign:
        | (Pick<GQL.ICampaign, '__typename' | 'id'> & {
              changesets: Pick<GQL.ICampaign['changesets'], 'nodes' | 'totalCount'>
              changesetPlans: Pick<GQL.ICampaign['changesetPlans'], 'nodes' | 'totalCount'>
          })
        | (Pick<GQL.ICampaignPlan, '__typename' | 'id'> & {
              changesets: Pick<GQL.ICampaignPlan['changesets'], 'nodes' | 'totalCount'>
          })
    persistLines: boolean
    campaignUpdates: Subject<void>
    changesetUpdates: Subject<void>

    history: H.History
    location: H.Location

    className?: string
}

const sumDiffStat = (nodes: (GQL.IExternalChangeset | GQL.IChangesetPlan)[], field: 'added' | 'deleted'): number =>
    nodes.reduce(
        (prev, next) =>
            prev + (next.diff ? next.diff.fileDiffs.diffStat[field] + next.diff.fileDiffs.diffStat.changed : 0),
        0
    )

/**
 * A tabbed view showing a campaign's or campaign plan's diffs and changesets.
 */
export const CampaignTabs: React.FunctionComponent<Props> = ({
    campaign,
    persistLines,
    history,
    location,
    className = '',
    isLightTheme,
    campaignUpdates,
    changesetUpdates,
}) => {
    const queryChangesetsConnection = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            campaign && campaign.__typename === 'CampaignPlan'
                ? queryChangesetPlans(campaign.id, args)
                : queryChangesets(campaign.id, args),
        [campaign]
    )

    const changesets = useMemo(
        () =>
            campaign.__typename === 'Campaign'
                ? [...campaign.changesets.nodes, ...campaign.changesetPlans.nodes]
                : campaign.changesets.nodes,
        [campaign]
    )
    const totalAdditions = useMemo(() => sumDiffStat(changesets, 'added'), [changesets])
    const totalDeletions = useMemo(() => sumDiffStat(changesets, 'deleted'), [changesets])

    return (
        <TabsWithLocalStorageViewStatePersistence
            storageKey="campaignTab"
            className={className}
            tabs={[
                {
                    id: 'diff',
                    label: (
                        <span className="e2e-campaign-diff-tab">
                            Diff <span className="text-success">+{totalAdditions}</span>{' '}
                            <span className="text-danger">-{totalDeletions}</span>
                        </span>
                    ),
                },
                {
                    id: 'changesets',
                    label: (
                        <span className="e2e-campaign-changesets-tab">
                            Changesets{' '}
                            <span className="badge badge-secondary badge-pill">
                                {campaign.changesets.totalCount +
                                    (campaign.__typename === 'Campaign' ? campaign.changesetPlans.totalCount : 0)}
                            </span>
                        </span>
                    ),
                },
            ]}
            tabClassName="tab-bar__tab--h5like"
        >
            <CampaignChangesets
                key="changesets"
                queryChangesetsConnection={queryChangesetsConnection}
                enablePublishing={campaign.__typename === 'Campaign'}
                campaignUpdates={campaignUpdates}
                changesetUpdates={changesetUpdates}
                history={history}
                location={location}
                className="mt-3"
                isLightTheme={isLightTheme}
            />
            <CampaignDiffs
                key="diff"
                queryChangesetsConnection={queryChangesetsConnection}
                changesetUpdates={changesetUpdates}
                persistLines={persistLines}
                history={history}
                location={location}
                className="mt-3"
                isLightTheme={isLightTheme}
            />
        </TabsWithLocalStorageViewStatePersistence>
    )
}
