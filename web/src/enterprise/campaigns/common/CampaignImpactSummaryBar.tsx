import { uniqBy } from 'lodash'
import AppsIcon from 'mdi-react/AppsIcon'
import ServerIcon from 'mdi-react/ServerIcon'
import React from 'react'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { SummaryCountBar, SummaryCountItemDescriptor } from '../../../components/summaryCountBar/SummaryCountBar'
import { DiffIcon } from '../../../util/octicons'
import { ChangesetOperationIcon } from '../../changesetsOLD/icons'
import { useCampaignFileDiffs } from '../detail/fileDiffs/useCampaignFileDiffs'
import { useCampaignRepositories } from '../detail/repositories/useCampaignRepositories'

interface Props {
    campaign: Pick<GQL.ICampaign, 'id' | 'rules' | 'isPreview'>

    className?: string
}

type Context = Pick<GQL.ICampaign, 'rules' | 'repositoryComparisons' | 'repositories' | 'isPreview'>

const countCampaignFilesChanged = (c: Pick<GQL.ICampaign, 'repositoryComparisons'>) =>
    c.repositoryComparisons.reduce((n, c) => n + (c.fileDiffs.totalCount || 0), 0)

const ITEMS: SummaryCountItemDescriptor<Context>[] = [
    {
        noun: 'rule applied',
        pluralNoun: 'rules applied',
        icon: ChangesetOperationIcon,
        count: c => JSON.parse(c.rules || '[]').length,
        condition: ({ isPreview }) => isPreview,
    },
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
    // TODO!(sqs): these are fake
    {
        noun: 'application affected',
        pluralNoun: 'applications affected',
        icon: AppsIcon,
        count: 5,
    },
    {
        noun: 'deployment group affected',
        pluralNoun: 'deployment groups affected',
        icon: ServerIcon,
        count: 3,
    },
]

const LOADING = 'loading' as const

/**
 * A bar that summarizes the contents and impact of a campaign.
 */
export const CampaignImpactSummaryBar: React.FunctionComponent<Props> = ({ campaign, className = '' }) => {
    // TODO(sqs): Make more efficient by not re-querying for these things (other components on the
    // page may also perform these queries, and the data can be shared among these components).
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

    const context: Context = {
        ...campaign,
        repositoryComparisons,
        repositories: uniqBy(repositoryComparisons.flatMap(c => c.headRepository), r => r.id),
    }
    return <SummaryCountBar<Context> className={className} context={context} itemDescriptors={ITEMS} />
}
