import React, { useCallback } from 'react'
import { Action, ActionType } from '../../../../../shared/src/api/types/action'
import { ChangesetCreationStatus } from '../../changesets/preview/backend'
import { CreateOrPreviewChangesetButton } from '../../tasks/list/item/CreateOrPreviewChangesetButton'

interface Props {
    /** The action. */
    action: ActionType['plan']

    /** Called when the button is clicked. */
    onClick: (action: Action, creationStatus: ChangesetCreationStatus) => void

    disabled?: boolean
    className?: string
}

/**
 * A button for a plan action that can be previewed and made into a changeset.
 */
export const PlanActionButton: React.FunctionComponent<Props> = ({ action, onClick, disabled, className = '' }) => {
    const onButtonClick = useCallback((creationStatus: ChangesetCreationStatus) => onClick(action, creationStatus), [
        action,
        onClick,
    ])
    return (
        <div className={`d-flex align-items-center ${className}`}>
            <CreateOrPreviewChangesetButton
                onClick={onButtonClick}
                disabled={disabled}
                className="mr-3"
                buttonClassName="btn-success"
            />
            <span className="text-muted">{action.plan.operations[0].command.title}</span>
        </div>
    )
}
