import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import SyncIcon from 'mdi-react/SyncIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'

const forceRefreshCampaign = (args: GQL.IForceRefreshCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation ForceRefreshCampaign($campaign: ID!) {
                forceRefreshCampaign(campaign: $campaign) {
                    id
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    className?: string
    buttonClassName?: string
}

/**
 * A button that force-refreshes a campaign.
 */
export const CampaignForceRefreshButton: React.FunctionComponent<Props> = ({
    campaign,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                await forceRefreshCampaign({ campaign: campaign.id })
                setIsLoading(false)
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error force-refreshing campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, campaign.id]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <SyncIcon className="icon-inline" />} Refresh
        </button>
    )
}
