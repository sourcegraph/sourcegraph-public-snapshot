import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ThreadsDropdownMenu } from '../../../threads/detail/sidebar/ThreadsDropdownMenu'

export const addThreadsToCampaign = (input: GQL.IAddThreadsToCampaignOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation AddThreadsToCampaign($campaign: ID!, $threads: [ID!]!) {
                addThreadsToCampaign(campaign: $campaign, threads: $threads) {
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
    onChange?: () => void

    className?: string
}

export const AddThreadToCampaignDropdownButton: React.FunctionComponent<Props> = ({
    campaign,
    onChange,
    className = '',
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onSelect = useCallback(
        async (thread: Pick<GQL.IThread, 'id'>) => {
            try {
                await addThreadsToCampaign({
                    campaign: campaign.id,
                    threads: [thread.id],
                })
                if (onChange) {
                    onChange()
                }
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error adding to campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [campaign.id, extensionsController.services.notifications.showMessages, onChange]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className} direction="down">
            <DropdownToggle color="" className="btn btn-primary">
                Add thread
            </DropdownToggle>
            <ThreadsDropdownMenu onSelect={onSelect} />
        </ButtonDropdown>
    )
}
