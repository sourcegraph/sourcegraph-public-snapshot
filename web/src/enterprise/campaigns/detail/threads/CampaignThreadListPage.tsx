import H from 'history'
import React, { useMemo } from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { useQueryParameter } from '../../../../util/useQueryParameter'
import { ThemeProps } from '../../../../theme'
import { CampaignAreaContext } from '../CampaignArea'
import { AddThreadToCampaignDropdownButton } from './AddThreadToCampaignDropdownButton'
import { CampaignThreadList } from './CampaignThreadList'
import { useCampaignThreads } from './useCampaignThreads'

interface Props
    extends Pick<CampaignAreaContext, 'campaign'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    className?: string

    location: H.Location
    history: H.History
}

export const CampaignThreadListPage: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => {
    const [query, onQueryChange, locationWithQuery] = useQueryParameter(props)
    const arg = useMemo(() => ({ filters: { query } }), [query])
    const [threads, onThreadsUpdate] = useCampaignThreads(campaign, arg)

    return (
        <div className={`campaign-thread-list-page ${className}`}>
            <CampaignThreadList
                {...props}
                threads={threads}
                onThreadsUpdate={onThreadsUpdate}
                campaign={campaign}
                query={query}
                onQueryChange={onQueryChange}
                locationWithQuery={locationWithQuery}
                action={<AddThreadToCampaignDropdownButton {...props} campaign={campaign} onChange={onThreadsUpdate} />}
            />
        </div>
    )
}
