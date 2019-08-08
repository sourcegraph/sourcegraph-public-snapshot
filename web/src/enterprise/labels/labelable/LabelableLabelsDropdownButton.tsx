import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { CampaignsDropdownMenu } from '../../campaigns/components/CampaignsDropdownMenu'
import { CampaignsIcon } from '../../campaigns/icons'
import { LabelIcon } from '../icons'

export const addLabelsToLabelable = (input: GQL.IAddLabelsToLabelableOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation AddLabelsToLabelable($labelable: ID!, $labels: [ID!]!) {
                addLabelsToLabelable(labelable: $labelable, labels: $labels) {
                    alwaysNil
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
    labelable: Pick<GQL.Labelable, 'id'>
    onChange?: () => void

    className?: string
    buttonClassName: string
}

/**
 * A dropdown button with a menu to add the thread to a labelable.
 */
export const ThreadCampaignsDropdownButton: React.FunctionComponent<Props> = ({
    labelable,
    onChange,
    className = '',
    buttonClassName,
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onSelect = useCallback(
        async (labelable: Pick<GQL.ICampaign, 'id'>) => {
            try {
                await addLabelsToLabelable({
                    labelable: labelable.id,
                    labels: [labelable.id],
                })
                if (onChange) {
                    onChange()
                }
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error updating labels: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, onChange]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle color="" className={buttonClassName}>
                <LabelIcon className="icon-inline small mr-2" /> Campaigns
            </DropdownToggle>
            <CampaignsDropdownMenu onSelect={onSelect} />
        </ButtonDropdown>
    )
}
