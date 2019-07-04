import React from 'react'
import { Action, ActionType, isActionType } from '../../../../../shared/src/api/types/action'
import { CommandActionButton } from '../../actions/internal/CommandActionButton'
import { PlanActionButton } from '../../actions/internal/PlanActionButton'
import { ChangesetCreationStatus } from '../../changesets/preview/backend'

interface Props {
    actions: readonly Action[]
    onPlanActionClick: (action: ActionType['plan'], creationStatus: ChangesetCreationStatus) => void
    onCommandActionClick: (action: ActionType['command']) => void

    disabled?: boolean
    className?: string
}

/**
 * The actions for a notification.
 */
export const NotificationActions: React.FunctionComponent<Props> = ({
    actions,
    onPlanActionClick,
    onCommandActionClick,
    disabled,
    className,
}) => {
    const planActions = actions.filter(isActionType('plan'))
    const commandActions = actions.filter(isActionType('command'))
    return (
        <div className={`d-flex flex-column align-items-start ${className}`}>
            {planActions.map((action, i) => (
                <PlanActionButton
                    key={i}
                    action={action}
                    onClick={onPlanActionClick}
                    disabled={disabled}
                    className="mr-3 mb-3"
                    buttonClassName="btn btn-success"
                />
            ))}
            {commandActions.length > 0 && (
                <div className="d-flex flex-wrap">
                    {commandActions.map((action, i) => (
                        <CommandActionButton
                            key={i}
                            action={action}
                            onClick={onCommandActionClick}
                            disabled={disabled}
                            className="btn btn-sm btn-secondary mr-3 mb-3"
                        />
                    ))}
                </div>
            )}
        </div>
    )
}
