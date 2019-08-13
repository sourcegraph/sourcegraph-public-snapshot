import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { useQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ParticipantList } from '../../../participants/ParticipantList'
import { CampaignAreaContext } from '../CampaignArea'
import { useCampaignParticipants } from './useCampaignParticipants'

interface Props extends Pick<CampaignAreaContext, 'campaign'>, ExtensionsControllerProps {
    className?: string

    location: H.Location
    history: H.History
}

export const CampaignParticipantListPage: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => {
    const [query, onQueryChange] = useQueryParameter(props)
    const participants = useCampaignParticipants(campaign)

    return (
        <div className={`campaign-participant-list-page ${className}`}>
            <ParticipantList {...props} participants={participants} query={query} onQueryChange={onQueryChange} />
        </div>
    )
}
