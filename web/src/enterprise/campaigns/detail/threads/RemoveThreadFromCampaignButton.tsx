import { NotificationType } from '@sourcegraph/extension-api-classes'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ThreadStateFields } from '../../../threads/common/threadState/threadState'

const removeThreadsFromCampaign = (input: GQL.IRemoveThreadsFromCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation RemoveThreadsFromCampaign($campaign: ID!, $threads: [ID!]!) {
                removeThreadsFromCampaign(campaign: $campaign, threads: $threads) {
                    alwaysNil
                }
            }
        `,
        input
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    thread: Pick<GQL.IThread, 'id' | 'repository' | 'title' | 'url' | 'externalURLs'> & ThreadStateFields
    onUpdate: () => void
}

/**
 * A button to remove a thread from a campaign.
 */
export const RemoveThreadFromCampaignButton: React.FunctionComponent<Props> = ({
    campaign,
    thread,
    onUpdate,
    extensionsController,
}) => {
    const onRemoveClick = useCallback(async () => {
        try {
            await removeThreadsFromCampaign({ campaign: campaign.id, threads: [thread.id] })
            onUpdate()
        } catch (err) {
            extensionsController.services.notifications.showMessages.next({
                message: `Error removing thread from campaign: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [campaign.id, extensionsController.services.notifications.showMessages, onUpdate, thread.id])

    return (
        <button className="btn btn-link btn-sm p-1" aria-label="Remove thread from campaign" onClick={onRemoveClick}>
            <CloseIcon className="icon-inline" />
        </button>
    )
}
