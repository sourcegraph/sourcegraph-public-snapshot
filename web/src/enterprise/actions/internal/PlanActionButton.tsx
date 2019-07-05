import React, { useCallback } from 'react'
import { ActionType } from '../../../../../shared/src/api/types/action'
import { ChangesetCreationStatus } from '../../changesets/preview/backend'
import {
    ChangesetButtonOrLink,
    ChangesetButtonOrLinkExistingChangeset,
} from '../../tasks/list/item/ChangesetButtonOrLink'

interface Props {
    /** The action. */
    action: ActionType['plan']

    /** Called when the button is clicked. */
    onClick: (action: ActionType['plan'], creationStatus: ChangesetCreationStatus) => void

    existingChangeset: ChangesetButtonOrLinkExistingChangeset

    disabled?: boolean
    className?: string
    buttonClassName?: string
}

/**
 * A label and button for a plan action that can be previewed and made into a changeset.
 */
export const PlanAction: React.FunctionComponent<Props> = ({
    action,
    onClick,
    existingChangeset,
    disabled,
    className = '',
    buttonClassName = '',
}) => {
    const onButtonClick = useCallback((creationStatus: ChangesetCreationStatus) => onClick(action, creationStatus), [
        action,
        onClick,
    ])
    return (
        <div className={`d-flex align-items-stretch ${className}`}>
            <span className="p-3 text-success border-left border-bottom border-top border-left-success border-bottom-success border-top-success">
                {action.plan.title}
            </span>
            <ChangesetButtonOrLink
                onClick={onButtonClick}
                existingChangeset={existingChangeset}
                disabled={disabled}
                buttonClassName={buttonClassName}
            />
        </div>
    )
}
