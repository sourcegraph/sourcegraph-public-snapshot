import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { LabelIcon } from '../icons'
import { LabelsDropdownMenu } from './LabelsDropdownMenu'

export const addLabelsToLabelable = (input: GQL.IAddLabelsToLabelableOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation AddLabelsToLabelable($labelable: ID!, $labels: [ID!]!) {
                addLabelsToLabelable(labelable: $labelable, labels: $labels) {
                    __typename
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
    repository: Pick<GQL.IRepository, 'id'>
    labelable: Pick<GQL.Labelable, 'id'>
    onChange?: () => void

    className?: string
    buttonClassName: string
}

/**
 * A dropdown button with a menu to add the thread to a labelable.
 */
export const LabelableLabelsDropdownButton: React.FunctionComponent<Props> = ({
    repository,
    labelable,
    onChange,
    className = '',
    buttonClassName,
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onSelect = useCallback(
        async (label: Pick<GQL.ILabel, 'id'>) => {
            try {
                await addLabelsToLabelable({
                    labelable: labelable.id,
                    labels: [label.id],
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
        [extensionsController.services.notifications.showMessages, labelable.id, onChange]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle color="" className={buttonClassName}>
                <LabelIcon className="icon-inline small mr-2" /> Labels
            </DropdownToggle>
            <LabelsDropdownMenu repository={repository} onSelect={onSelect} />
        </ButtonDropdown>
    )
}
