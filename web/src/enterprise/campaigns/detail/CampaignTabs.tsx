import React, { useMemo } from 'react'
import H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../shared/src/theme'
import { TabsWithLocalStorageViewStatePersistence } from '../../../../../shared/src/components/Tabs'
import { CampaignDiffs } from './diffs/CampaignDiffs'
import { CampaignChangesets } from './changesets/CampaignChangesets'

interface Props extends ThemeProps {
    changesets: Pick<GQL.IExternalChangesetConnection | GQL.IChangesetPlanConnection, 'nodes' | 'totalCount'>
    persistLines: boolean

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
    changesets,
    persistLines,
    history,
    location,
    className = '',
    isLightTheme,
}) => {
    const totalAdditions = useMemo(() => sumDiffStat(changesets.nodes, 'added'), [changesets.nodes])
    const totalDeletions = useMemo(() => sumDiffStat(changesets.nodes, 'deleted'), [changesets.nodes])

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
                            Changesets <span className="badge badge-secondary badge-pill">{changesets.totalCount}</span>
                        </span>
                    ),
                },
            ]}
            tabClassName="tab-bar__tab--h5like"
        >
            <CampaignChangesets
                key="changesets"
                changesets={changesets}
                history={history}
                location={location}
                className="mt-3"
                isLightTheme={isLightTheme}
            />
            <CampaignDiffs
                key="diff"
                changesets={changesets}
                persistLines={persistLines}
                history={history}
                location={location}
                className="mt-3"
                isLightTheme={isLightTheme}
            />
        </TabsWithLocalStorageViewStatePersistence>
    )
}
