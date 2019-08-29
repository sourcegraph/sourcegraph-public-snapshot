import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'

const publishDraftCampaign = (args: GQL.IPublishDraftCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation PublishDraftCampaign($campaign: ID!) {
                publishDraftCampaign(campaign: $campaign) {
                    id
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    onComplete?: () => void
    className?: string
    buttonClassName?: string
}

/**
 * A button that publishes a draft campaign.
 */
export const PublishDraftCampaignButton: React.FunctionComponent<Props> = ({
    campaign,
    onComplete,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            if (
                !confirm(
                    'Are you sure? Publishing the campaign will create branches, issues, changesets, and notifications.'
                )
            ) {
                return
            }
            setIsLoading(true)
            try {
                await publishDraftCampaign({ campaign: campaign.id })
                setIsLoading(false)
                if (onComplete) {
                    onComplete()
                }
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error publishing draft campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, onComplete, campaign.id]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading && <LoadingSpinner className="icon-inline" />} Publish campaign
        </button>
    )
}
