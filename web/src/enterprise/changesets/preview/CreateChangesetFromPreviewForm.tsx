import H from 'history'
import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { updateThread } from '../../../discussions/backend'
import { ThreadSettings } from '../../threads/settings'

interface Props extends ExtensionsControllerNotificationProps {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    className?: string
    history: H.History
}

/**
 * A form to create a non-preview changeset from a preview changeset.
 */
export const CreateChangesetFromPreviewForm: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    className = '',
    history,
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)

    const [uncommittedTitle, setUncommittedTitle] = useState(thread.title)
    const onChangeTitle = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setUncommittedTitle(e.currentTarget.value)
    }, [])

    const onSubmit: React.FormEventHandler = useCallback(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const updatedThread = await updateThread({
                    threadID: thread.id,
                    title: uncommittedTitle,
                    status: GQL.ThreadStatus.OPEN_ACTIVE,
                })
                setIsLoading(false)
                onThreadUpdate(updatedThread)
                history.push(updatedThread.url)
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error creating changeset: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, history, onThreadUpdate, thread.id, uncommittedTitle]
    )

    return (
        <Form className={className} onSubmit={onSubmit}>
            <div className="form-group">
                <input
                    type="text"
                    className="form-control"
                    value={uncommittedTitle}
                    onChange={onChangeTitle}
                    placeholder="Title"
                    autoComplete="off"
                    autoFocus={true}
                    disabled={isLoading}
                />
            </div>
            <div className="form-group d-flex">
                <textarea
                    className="form-control"
                    onChange={e => {
                        alert('not implemented')
                    }}
                    placeholder="Description"
                    style={{ resize: 'vertical', minHeight: '150px' }}
                    disabled={isLoading}
                />
            </div>
            <div className="d-flex justify-content-end">
                <button type="submit" className="btn btn-lg btn-success" disabled={isLoading}>
                    Create changeset
                </button>
            </div>
        </Form>
    )
}
