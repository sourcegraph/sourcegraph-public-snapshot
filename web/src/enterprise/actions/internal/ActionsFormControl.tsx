import React from 'react'
import { Action, isActionType } from '../../../../../shared/src/api/types/action'
import { ChangesetCreationStatus } from '../../changesets/preview/backend'
import { CommandActionButton } from './CommandActionButton'
import { PlanActionButton } from './PlanActionButton'

interface Props {
    actions: readonly Action[]
    onActionClick: (action: Action, creationStatus?: ChangesetCreationStatus) => void

    className?: string
    buttonClassName?: string
}

/**
 * A form control that displays {@link sourcegraph.Action}s.
 *
 * TODO!(sqs): dedupe with ThreadInboxItemActions?
 */
export const ActionsFormControl: React.FunctionComponent<Props> = ({
    actions,
    onActionClick,
    className,
    buttonClassName = 'btn btn-link text-decoration-none',
}) => {
    const planActions = actions.filter(isActionType('plan'))
    const commandActions = actions.filter(isActionType('command'))
    return (
        <div className={`d-flex flex-column align-items-start ${className}`}>
            {planActions.map((action, i) => (
                <PlanActionButton key={i} action={action} onClick={onActionClick} className="mb-2" />
            ))}
            {commandActions.length > 0 && (
                <div className="d-flex flex-wrap">
                    {commandActions.map((action, i) => (
                        <CommandActionButton
                            key={i}
                            action={action}
                            onClick={onActionClick}
                            className={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
                        />
                    ))}
                </div>
            )}
        </div>
    )
}
