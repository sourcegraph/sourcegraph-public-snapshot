import CommentTextMultipleIcon from 'mdi-react/CommentTextMultipleIcon'
import React from 'react'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { SummaryCountBar, SummaryCountItemDescriptor } from '../../../components/summaryCountBar/SummaryCountBar'
import { DiffStat } from '../../../repo/compare/DiffStat'
import { DiffIcon, GitPullRequestIcon } from '../../../util/octicons'
import { DiagnosticsIcon } from '../../checks/icons'
import { CampaignImpactSummary, useCampaignImpactSummary } from './useCampaignImpactSummary'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id' | 'url'>

    className?: string
}

interface Context extends CampaignImpactSummary {
    campaign?: Pick<GQL.ICampaign, 'url'>
}

const ITEMS: SummaryCountItemDescriptor<Context>[] = [
    {
        noun: 'discussion',
        icon: CommentTextMultipleIcon,
        count: c => c.discussions,
        condition: c => c.discussions > 0,
        url: c => c.campaign && c.campaign.url,
    },
    {
        noun: 'issue',
        icon: DiagnosticsIcon,
        count: c => c.issues,
        condition: c => c.issues > 0,
        url: c => c.campaign && `${c.campaign.url}/threads`,
    },
    {
        noun: 'changeset',
        icon: GitPullRequestIcon,
        count: c => c.changesets,
        condition: c => c.changesets > 0,
        url: c => c.campaign && `${c.campaign.url}/threads`,
    },
    {
        noun: 'repository affected',
        pluralNoun: 'repositories affected',
        icon: RepositoryIcon,
        count: c => c.repositories,
        url: c => c.campaign && `${c.campaign.url}/changes`,
    },
    {
        noun: 'file changed',
        pluralNoun: 'files changed',
        icon: DiffIcon,
        count: c => c.files,
        url: c => c.campaign && `${c.campaign.url}/changes`,
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
    return <CampaignImpactSummaryBarNoFetch campaign={campaign} impactSummary={impactSummary} />
}

export const CampaignImpactSummaryBarNoFetch: React.FunctionComponent<
    Pick<Props, Exclude<keyof Props, 'campaign'>> & {
        campaign?: Pick<GQL.ICampaign, 'id' | 'url'>
        impactSummary: CampaignImpactSummary
    }
> = ({ campaign, impactSummary, className = '' }) => (
    <SummaryCountBar<Context> className={className} context={{ ...impactSummary, campaign }} itemDescriptors={ITEMS} />
)
