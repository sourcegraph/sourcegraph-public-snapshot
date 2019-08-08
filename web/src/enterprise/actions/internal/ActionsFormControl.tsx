import { NotificationType } from '@sourcegraph/extension-api-classes'
import React, { useState } from 'react'
import { Action, isCommandOnlyAction } from '../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ActionRadioButton } from './ActionRadioButton'
import { CommandActionButton } from './CommandActionButton'

interface Props extends ExtensionsControllerProps {
    actions: readonly Action[]
    selectedAction: Action | null
    onActionSetSelected: (value: boolean, action: Action) => void

    className?: string
    buttonClassName?: string
    activeButtonClassName?: string
    inactiveButtonClassName?: string
}

/**
 * A form control that displays {@link sourcegraph.Action}s.
 */
export const ActionsFormControl: React.FunctionComponent<Props> = ({
    actions,
    selectedAction,
    onActionSetSelected,
    className,
    buttonClassName = '',
    activeButtonClassName = '',
    inactiveButtonClassName = '',
    extensionsController,
}) => {
    const planActions = actions.filter(action => !isCommandOnlyAction(action))
    const commandActions = actions.filter(isCommandOnlyAction)

    const [isLoading, setIsLoading] = useState(false)

    return (
        <div className={`d-flex flex-column align-items-start ${className}`}>
            {planActions.map((action, i) => (
                <ActionRadioButton
                    key={i}
                    action={action}
                    onChange={onActionSetSelected}
                    className="mr-2 mb-2"
                    buttonClassName={buttonClassName}
                    activeButtonClassName={activeButtonClassName}
                    inactiveButtonClassName={inactiveButtonClassName}
                    value={selectedAction === action}
                    disabled={isLoading}
                />
            ))}
            {commandActions.length > 0 && (
                <div className="d-flex flex-wrap">
                    {commandActions.map((action, i) => (
                        <CommandActionButton
                            key={i}
                            action={action}
                            disabled={isLoading}
                            // tslint:disable-next-line: jsx-no-lambda
                            onClick={async () => {
                                setIsLoading(true)
                                try {
                                    await extensionsController.executeCommand(action.command)
                                    setIsLoading(false)
                                } catch (err) {
                                    setIsLoading(false)
                                    extensionsController.services.notifications.showMessages.next({
                                        message: `Error: ${err.message}`,
                                        type: NotificationType.Error,
                                    })
                                }
                            }}
                            className={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
                        />
                    ))}
                </div>
            )}
        </div>
    )
}
