import { CodeAction } from '@sourcegraph/extension-api-types'
import React from 'react'
import { Action, isActionType } from '../../../../../shared/src/api/types/action'
import { ChangesetCreationStatus } from '../../changesets/preview/backend'
import { ActionRadioButton } from './ActionRadioButton'
import { CommandActionButton } from './CommandActionButton'
import { PlanAction } from './PlanActionButton'

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
                // TODO!(sqs) <PlanAction key={i} action={action} onClick={onActionClick} className="mb-2" />
                <ActionRadioButton
                    key={i}
                    action={{ title: action.title } as CodeAction /* TODO!(sqs) */}
                    onChange={() => alert('TODO!(sqs)')}
                    buttonClassName="btn btn-primary"
                    className={`mr-2 mb-2`}
                    value={false}
                />
            ))}
            {commandActions.length > 0 && (
                <div className="d-flex flex-wrap">
                    {commandActions.map((action, i) => (
                        <CommandActionButton
                            key={i}
                            action={action}
                            onClick={onActionClick}
                            className={`${buttonClassName} mr-2 mb-2`}
                        />
                    ))}
                </div>
            )}
        </div>
    )
}
