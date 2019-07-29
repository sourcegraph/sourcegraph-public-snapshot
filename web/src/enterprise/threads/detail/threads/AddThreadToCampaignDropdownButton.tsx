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

const addThreadsToThread = (input: GQL.IAddThreadsToThreadOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation AddThreadsToThread($thread: ID!, $threads: [ID!]!) {
                addThreadsToThread(thread: $thread, threads: $threads) {
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
    thread: Pick<GQL.IThread, 'id'>
    onAdd: () => void

    className?: string
}

export const AddThreadToThreadDropdownButton: React.FunctionComponent<Props> = ({
    thread,
    onAdd,
    className = '',
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onThreadAdd = useCallback(
        async (thread: Pick<GQL.IDiscussionThread, 'id'>) => {
            try {
                await addThreadsToThread({
                    thread: thread.id,
                    threads: [thread.id],
                })
                onAdd()
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error adding thread to thread: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [thread.id, extensionsController.services.notifications.showMessages, onAdd]
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
