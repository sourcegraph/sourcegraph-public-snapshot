import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback } from 'react'
import { CodeAction } from 'sourcegraph'
import { ActionButton } from './ActionButton'
import { ActionRadioButton } from './ActionRadioButton'
import { ActionsDropdownButton } from './ActionsDropdownButton'

interface Props {
    actions: readonly CodeAction[]
    activeAction: CodeAction | undefined
    onActionSetActive: (activeAction: CodeAction | undefined) => void
    onActionClick: (action: CodeAction) => void

    className?: string
    buttonClassName?: string
    inactiveButtonClassName?: string
    activeButtonClassName?: string
}

/**
 * A form control that displays {@link sourcegraph.CodeAction}s.
 *
 * TODO!(sqs): dedupe with ThreadInboxItemActions?
 */
export const ActionsFormControl: React.FunctionComponent<Props> = ({
    actions,
    activeAction,
    onActionClick,
    onActionSetActive,
    className,
    buttonClassName = 'btn btn-link text-decoration-none',
    inactiveButtonClassName,
    activeButtonClassName,
}) => {
    const actionsWithEdit = actions.filter(({ edit }) => !!edit)
    const actionsWithCommand = actions.filter(
        ({ diagnostics, command }) => !!command && diagnostics && diagnostics.length > 0
    )
    const secondaryCodeActions = actions.filter(
        ({ diagnostics, command }) => !!command && (!diagnostics || diagnostics.length === 0)
    )

    const onActionValueChange = useCallback(
        (value: boolean, action: CodeAction) => {
            onActionSetActive(value ? action : undefined)
        },
        [onActionSetActive]
    )

    return (
        <div className={`task-list-item-actions d-flex ${className}`}>
            {actionsWithEdit.map((action, i) => (
                <ActionRadioButton
                    key={i}
                    action={action}
                    value={action === activeAction}
                    onChange={onActionValueChange}
                    className="mb-2"
                    buttonClassName={buttonClassName}
                    activeButtonClassName={activeButtonClassName}
                    inactiveButtonClassName={inactiveButtonClassName}
                />
            ))}
            {actionsWithCommand.map((action, i) => (
                <ActionButton
                    key={i}
                    action={action}
                    onClick={onActionClick}
                    className={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
                />
            ))}
            <ActionButton
                action={{ title: 'Preview and edit changeset...' }}
                onClick={onActionClick}
                className={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
            />
            {secondaryCodeActions.length > 0 && (
                <ActionsDropdownButton
                    actions={secondaryCodeActions}
                    buttonClassName={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
                />
            )}
        </div>
    )
}
