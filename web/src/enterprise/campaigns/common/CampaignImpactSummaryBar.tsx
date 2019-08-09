import React from 'react'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { SummaryCountBar, SummaryCountItemDescriptor } from '../../../components/summaryCountBar/SummaryCountBar'
import { DiffStat } from '../../../repo/compare/DiffStat'
import { DiffIcon } from '../../../util/octicons'
import { CampaignImpactSummary, useCampaignImpactSummary } from './useCampaignImpactSummary'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
}

const ITEMS: SummaryCountItemDescriptor<CampaignImpactSummary>[] = [
    {
        noun: 'repository affected',
        pluralNoun: 'repositories affected',
        icon: RepositoryIcon,
        count: c => c.repositories,
    },
    {
        noun: 'file changed',
        pluralNoun: 'files changed',
        icon: DiffIcon,
        count: c => c.files,
        after: c => <DiffStat {...c.diffStat} expandedCounts={true} className="d-inline-flex ml-3" />,
    },
]

const LOADING = 'loading' as const

/**
 * A bar that summarizes the contents and impact of a campaign.
 */
export const CampaignImpactSummaryBar: React.FunctionComponent<Props> = ({ campaign, className = '' }) => {
    const impactSummary = useCampaignImpactSummary(campaign)
    if (impactSummary === LOADING || isErrorLike(impactSummary)) {
        return null
    }
    return (
        <SummaryCountBar<CampaignImpactSummary> className={className} context={impactSummary} itemDescriptors={ITEMS} />
    )
}
