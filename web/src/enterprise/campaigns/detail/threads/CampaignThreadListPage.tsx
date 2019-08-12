import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { useQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { CampaignAreaContext } from '../CampaignArea'
import { AddThreadToCampaignDropdownButton } from './AddThreadToCampaignDropdownButton'
import { CampaignThreadList } from './CampaignThreadList'
import { useCampaignThreads } from './useCampaignThreads'

interface Props extends Pick<CampaignAreaContext, 'campaign'>, ExtensionsControllerProps {
    className?: string

    location: H.Location
    history: H.History
}

export const CampaignThreadListPage: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => {
    const [query, onQueryChange] = useQueryParameter(props)
    const [threads, onThreadsUpdate] = useCampaignThreads(campaign)

    return (
        <div className={`campaign-thread-list-page ${className}`}>
            <CampaignThreadList
                {...props}
                threads={threads}
                campaign={campaign}
                query={query}
                onQueryChange={onQueryChange}
                action={<AddThreadToCampaignDropdownButton {...props} campaign={campaign} onChange={onThreadsUpdate} />}
            />
        </div>
    )
}
