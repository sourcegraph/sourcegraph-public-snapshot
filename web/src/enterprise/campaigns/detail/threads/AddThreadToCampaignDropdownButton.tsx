import PlusBoxIcon from 'mdi-react/PlusBoxIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ThreadDropdownMenu } from './ThreadDropdownMenu'

const addThreadsToCampaign = (input: GQL.IAddThreadsToCampaignOnCampaignsMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation AddThreadsToCampaign($campaign: ID!, $threads: [ID!]!) {
                campaigns {
                    addThreadsToCampaign(campaign: $campaign, threads: $threads) {
                        alwaysNil
                    }
                }
            }
        `,
        input
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    onAdd: () => void

    className?: string
}

export const AddThreadToCampaignDropdownButton: React.FunctionComponent<Props> = ({
    campaign,
    onAdd,
    className = '',
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onThreadAdd = useCallback(
        async (thread: Pick<GQL.IDiscussionThread, 'id'>) => {
            try {
                await addThreadsToCampaign({
                    campaign: campaign.id,
                    threads: [thread.id],
                })
                onAdd()
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error adding thread to campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [campaign.id, extensionsController.services.notifications.showMessages, onAdd]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle color="" className="btn btn-primary">
                <PlusBoxIcon className="icon-inline mr-2" /> Add thread
            </DropdownToggle>
            <ThreadDropdownMenu onThreadClick={onThreadAdd} />
        </ButtonDropdown>
    )
}
