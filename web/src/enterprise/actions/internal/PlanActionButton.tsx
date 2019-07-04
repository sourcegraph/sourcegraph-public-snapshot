import React, { useCallback } from 'react'
import { ActionType } from '../../../../../shared/src/api/types/action'
import { ChangesetCreationStatus } from '../../changesets/preview/backend'
import { CreateOrPreviewChangesetButton } from '../../tasks/list/item/CreateOrPreviewChangesetButton'

interface Props {
    /** The action. */
    action: ActionType['plan']

    /** Called when the button is clicked. */
    onClick: (action: ActionType['plan'], creationStatus: ChangesetCreationStatus) => void

    disabled?: boolean
    className?: string
    buttonClassName?: string
}

/**
 * A button for a plan action that can be previewed and made into a changeset.
 */
export const PlanActionButton: React.FunctionComponent<Props> = ({
    action,
    onClick,
    disabled,
    className = '',
    buttonClassName = '',
}) => {
    const onButtonClick = useCallback((creationStatus: ChangesetCreationStatus) => onClick(action, creationStatus), [
        action,
        onClick,
    ])
    return (
        <div className={`d-flex flex-column ${className} p-3 border border-success`}>
            <span className="text-success mb-3">
                <strong>Fix:</strong> {action.plan.operations[0].command.title}
            </span>
            <CreateOrPreviewChangesetButton
                onClick={onButtonClick}
                disabled={disabled}
                className=""
                buttonClassName={buttonClassName}
            />
        </div>
    )
}
