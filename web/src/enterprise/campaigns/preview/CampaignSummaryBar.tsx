import { uniqBy } from 'lodash'
import React from 'react'
import { Link } from 'react-router-dom'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { DiffIcon } from '../../../util/octicons'
import { useCampaignFileDiffs } from '../detail/fileDiffs/useCampaignFileDiffs'
import { useCampaignRepositories } from '../detail/repositories/useCampaignRepositories'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
}

interface SummaryItem {
    noun: string
    pluralNoun?: string
    icon: React.ComponentType<{ className?: string }>
    count: number | ((campaign: Pick<GQL.ICampaign, 'repositoryComparisons' | 'repositories'>) => number) | null
}

export const countCampaignFilesChanged = (c: Pick<GQL.ICampaign, 'repositoryComparisons'>) =>
    c.repositoryComparisons.reduce((n, c) => n + (c.fileDiffs.totalCount || 0), 0)

const ITEMS: SummaryItem[] = [
    {
        noun: 'repository affected',
        pluralNoun: 'repositories affected',
        icon: RepositoryIcon,
        count: c => c.repositories.length,
    },
    {
        noun: 'file changed',
        pluralNoun: 'files changed',
        icon: DiffIcon,
        count: countCampaignFilesChanged,
    },
]

const LOADING = 'loading' as const

/**
 * A bar that summarizes the contents and impact of a campaign.
 */
export const CampaignSummaryBar: React.FunctionComponent<Props> = ({ campaign, className = '' }) => {
    // TODO!(sqs) make more efficient by not re-querying graphql for these things because the parent
    // CampaignPreviewPage already performs these queries
    const repositoryComparisons = useCampaignFileDiffs(campaign)
    const repositories = useCampaignRepositories(campaign)

    if (
        repositoryComparisons === LOADING ||
        isErrorLike(repositoryComparisons) ||
        repositories === LOADING ||
        isErrorLike(repositories)
    ) {
        return null
    }

    // TODO!(sqs)
    const campaignInfo = {
        repositoryComparisons,
        repositories: uniqBy(repositoryComparisons.flatMap(c => c.headRepository), r => r.id),
    }

    return (
        <nav className={`campaign-summary-bar border ${className}`}>
            <ul className="nav w-100">
                {ITEMS.map(({ icon: Icon, ...item }, i) => {
                    const count =
                        typeof item.count === 'number'
                            ? item.count
                            : typeof item.count === 'function'
                            ? item.count(campaignInfo)
                            : null
                    return (
                        <li key={i} className="nav-item flex-1 text-center">
                            <Link to="TODO!(sqs)" className="nav-link">
                                <Icon className="icon-inline text-muted" /> <strong>{count}</strong>{' '}
                                <span className="text-muted">{pluralize(item.noun, count || 0, item.pluralNoun)}</span>
                            </Link>
                        </li>
                    )
                })}
            </ul>
        </nav>
    )
}
