import React from 'react'
import { CodeAction } from 'sourcegraph'
import { ThreadInboxItemActionsDropdownButton } from '../../../threads/detail/inbox/item/actions/ThreadInboxItemActionsDropdownButton'

interface Props {
    codeActions: CodeAction[]
    onCodeActionClick: (codeAction: CodeAction) => void

    className?: string
    buttonClassName?: string
}

/**
 * The actions that can be performed on a task.
 *
 * TODO!(sqs): dedupe with ThreadInboxItemActions?
 */
export const TasksListItemActions: React.FunctionComponent<Props> = ({
    codeActions,
    onCodeActionClick: onCodeActionClick,
    className,
    buttonClassName = 'btn btn-link text-decoration-none',
}) => {
    const primaryCodeActions = codeActions.filter(({ diagnostics }) => !!diagnostics && diagnostics.length > 0)
    const secondaryCodeActions = codeActions.filter(
        ({ diagnostics, command }) => !!command && (!diagnostics || diagnostics.length === 0)
    )

    return (
        <div className={className}>
            {primaryCodeActions.length > 0 && (
                <ul className="list-unstyled mb-0 d-flex flex-column flex-wrap">
                    {primaryCodeActions.map((codeAction, i) => (
                        <li key={i}>
                            <button
                                type="button"
                                className={`${buttonClassName} mr-2 mb-2`}
                                // tslint:disable-next-line: jsx-no-lambda
                                onClick={() => onCodeActionClick(codeAction)}
                            >
                                {codeAction.title}
                            </button>
                        </li>
                    ))}
                </ul>
            )}
            {secondaryCodeActions.length > 0 && (
                <ThreadInboxItemActionsDropdownButton
                    codeActions={secondaryCodeActions}
                    buttonClassName={`${buttonClassName} mr-2 mb-2`}
                />
            )}
        </div>
    )
}
