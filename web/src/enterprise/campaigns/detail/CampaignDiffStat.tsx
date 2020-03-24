import React, { useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { DiffStat } from '../../../components/diff/DiffStat'

interface NodesWithDiffStat {
    nodes: {
        diff: {
            fileDiffs: { diffStat: { added: number; changed: number; deleted: number } }
        } | null
    }[]
}

export interface CampaignDiffstatProps {
    campaign?: Pick<GQL.ICampaign, '__typename'> & {
        changesets: NodesWithDiffStat
        patches: NodesWithDiffStat
    }
    patchSet?: Pick<GQL.IPatchSet, '__typename'> & {
        patches: NodesWithDiffStat
    }

    className?: string
}

const sumDiffStat = (nodes: NodesWithDiffStat['nodes'], field: 'added' | 'changed' | 'deleted'): number =>
    nodes.reduce((prev, next) => prev + (next.diff ? next.diff.fileDiffs.diffStat[field] : 0), 0)

/**
 * Total diff stat of a campaign or patchset, including all changesets and patches
 */
export const CampaignDiffStat: React.FunctionComponent<CampaignDiffstatProps> = ({ campaign, patchSet, className }) => {
    const changesets = useMemo(
        () => (campaign ? [...campaign.changesets.nodes, ...campaign.patches.nodes] : patchSet!.patches.nodes),
        [campaign, patchSet]
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
