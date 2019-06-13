import { NotificationType } from '@sourcegraph/extension-api-classes'
import { useCallback, useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { ActionType } from '../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'

type Callback = (action: sourcegraph.Action | ActionType['command']) => void

/**
 * A React hook that returns an `onActionClick` callback that executes an action, displaying a
 * notification if execution fails.
 */
export const useOnActionClickCallback = (
    extensionsController: ExtensionsControllerProps['extensionsController']
): [Callback, boolean /* isLoading */] => {
    const [isLoading, setIsLoading] = useState(false)
    const callback = useCallback(
        async (action: sourcegraph.Action | ActionType['command']) => {
            try {
                const command = 'command' in action ? action.command : undefined
                if (command) {
                    setIsLoading(true)
                    await extensionsController.executeCommand(command)
                    setIsLoading(false)
                }
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error running action: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController, setIsLoading]
    )
    return [callback, isLoading]
}
