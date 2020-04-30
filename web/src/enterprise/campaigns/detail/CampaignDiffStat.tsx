import React, { useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { DiffStat } from '../../../components/diff/DiffStat'

export interface CampaignDiffstatProps {
    campaign?: Pick<GQL.ICampaign, '__typename'> & {
        diffStat: { added: number; changed: number; deleted: number }
    }
    patchSet?: Pick<GQL.IPatchSet, '__typename'> & {
        diffStat: { added: number; changed: number; deleted: number }
    }

    className?: string
}

/**
 * Total diff stat of a campaign or patchset, including all changesets and patches
 */
export const CampaignDiffStat: React.FunctionComponent<CampaignDiffstatProps> = ({ campaign, patchSet, className }) => {
    const { added, changed, deleted } = useMemo(() => (campaign ? campaign.diffStat : patchSet!.diffStat), [
        campaign,
        patchSet,
    ])

    if (added + changed + deleted === 0) {
        return <></>
    }

    return (
        <div className={className}>
            <DiffStat expandedCounts={true} added={added} changed={changed} deleted={deleted} />
        </div>
    )
}
