import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { CampaignsDropdownMenu } from '../../../campaigns/components/CampaignsDropdownMenu'
import { CampaignsIcon } from '../../../campaigns/icons'

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
    thread: Pick<GQL.IThread, 'id'>
    onChange?: () => void

    className?: string
    buttonClassName: string
}

/**
 * A dropdown button with a menu to add the thread to a campaign.
 */
export const ThreadCampaignsDropdownButton: React.FunctionComponent<Props> = ({
    thread,
    onChange,
    className = '',
    buttonClassName,
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onSelect = useCallback(
        async (campaign: Pick<GQL.ICampaign, 'id'>) => {
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
        [extensionsController.services.notifications.showMessages, onChange, thread.id]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle color="" className={buttonClassName}>
                <CampaignsIcon className="icon-inline small mr-2" /> Campaigns
            </DropdownToggle>
            <CampaignsDropdownMenu onSelect={onSelect} />
        </ButtonDropdown>
    )
}
