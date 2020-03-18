import React, { useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { DiffStat } from '../../../components/diff/DiffStat'

export interface CampaignDiffstatProps {
    campaign:
        | (Pick<GQL.ICampaign, '__typename'> & {
              changesets: Pick<GQL.ICampaign['changesets'], 'nodes'>
              changesetPlans: Pick<GQL.ICampaign['changesetPlans'], 'nodes'>
          })
        | (Pick<GQL.ICampaignPlan, '__typename'> & {
              changesetPlans: Pick<GQL.ICampaignPlan['changesetPlans'], 'nodes'>
          })

    className?: string
}

const sumDiffStat = (
    nodes: (GQL.IExternalChangeset | GQL.IChangesetPlan)[],
    field: 'added' | 'changed' | 'deleted'
): number => nodes.reduce((prev, next) => prev + (next.diff ? next.diff.fileDiffs.diffStat[field] : 0), 0)

/**
 * The status of a campaign's jobs, plus its closed state and errors.
 */
export const CampaignDiffStat: React.FunctionComponent<CampaignDiffstatProps> = ({ campaign, className }) => {
    const changesets = useMemo(
        () =>
            campaign.__typename === 'Campaign'
                ? [...campaign.changesets.nodes, ...campaign.changesetPlans.nodes]
                : campaign.changesetPlans.nodes,
        [campaign]
    )
    const added = useMemo(() => sumDiffStat(changesets, 'added'), [changesets])
    const changed = useMemo(() => sumDiffStat(changesets, 'changed'), [changesets])
    const deleted = useMemo(() => sumDiffStat(changesets, 'deleted'), [changesets])

    if (added + changed + deleted === 0) {
        return <></>
    }

    return (
        <div className={className}>
            <DiffStat expandedCounts={true} added={added} changed={changed} deleted={deleted} />
        </div>
    )
}
