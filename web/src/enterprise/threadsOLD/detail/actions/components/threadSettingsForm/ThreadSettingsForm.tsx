import { NotificationType } from '@sourcegraph/extension-api-classes'
import React, { useCallback, useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../../shared/src/graphql/schema'
import { Form } from '../../../../../../components/Form'
import { updateThread } from '../../../../../../discussions/backend'
import { ThreadSettings } from '../../../../settings'

interface CommonProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'title' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
}

interface RenderChildrenProps extends CommonProps {
    uncommittedSettings: ThreadSettings
    setUncommittedSettings: (value: ThreadSettings) => void
    isLoading: boolean
}

interface Props extends CommonProps, ExtensionsControllerProps {
    children: (props: RenderChildrenProps) => JSX.Element | null
    initialUncommittedSettings?: ThreadSettings
}

/**
 * A form that updates values in the thread settings based on form field inputs.
 */
export const ThreadSettingsForm: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    initialUncommittedSettings,
    children,
    extensionsController,
}) => {
    const [uncommittedSettings, setUncommittedSettings] = useState<ThreadSettings>(
        initialUncommittedSettings || threadSettings
    )

    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const updatedThread = await updateThread({
                    threadID: thread.id,
                    settings: JSON.stringify(uncommittedSettings, null, 2),
                })
                setIsLoading(false)
                onThreadUpdate(updatedThread)
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error saving: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [thread.id, uncommittedSettings, onThreadUpdate, extensionsController.services.notifications.showMessages]
    )

    return (
        <Form className="form" onSubmit={onSubmit}>
            {children({
                thread,
                onThreadUpdate,
                threadSettings,
                isLoading,
                uncommittedSettings,
                setUncommittedSettings,
            })}
        </Form>
    )
}
