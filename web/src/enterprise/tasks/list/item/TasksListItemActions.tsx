import CloseIcon from 'mdi-react/CloseIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React from 'react'
import { CodeAction } from 'sourcegraph'
import { ThreadInboxItemActionsDropdownButton } from '../../../threads/detail/inbox/item/actions/ThreadInboxItemActionsDropdownButton'

interface Props {
    codeActions: CodeAction[]
    activeCodeAction: CodeAction | undefined
    onCodeActionSetActive: (activeCodeAction: CodeAction | undefined) => void
    onCodeActionClick: (codeAction: CodeAction) => void

    className?: string
    buttonClassName?: string
    inactiveButtonClassName?: string
    activeButtonClassName?: string
}

/**
 * The actions that can be performed on a task.
 *
 * TODO!(sqs): dedupe with ThreadInboxItemActions?
 */
export const TasksListItemActions: React.FunctionComponent<Props> = ({
    codeActions,
    activeCodeAction,
    onCodeActionClick,
    onCodeActionSetActive,
    className,
    buttonClassName = 'btn btn-link text-decoration-none',
    inactiveButtonClassName,
    activeButtonClassName,
}) => {
    // TODO!(sqs)
    // const primaryCodeActions = codeActions.filter(({ diagnostics }) => !!diagnostics && diagnostics.length > 0)
    // const secondaryCodeActions = codeActions.filter(
    //     ({ diagnostics, command }) => !!command && (!diagnostics || diagnostics.length === 0)
    // )

    const codeActionsWithEdit = codeActions.filter(({ edit }) => !!edit)
    const codeActionsWithCommand = codeActions.filter(
        ({ diagnostics, command }) => !!command && diagnostics && diagnostics.length > 0
    )
    const secondaryCodeActions = codeActions.filter(
        ({ diagnostics, command }) => !!command && (!diagnostics || diagnostics.length === 0)
    )

    return (
        <div className={`task-list-item-actions d-flex flex-column ${className}`}>
            {codeActionsWithEdit.length > 0 && (
                <div className="d-flex flex-column align-items-start" data-toggle="buttons">
                    {codeActionsWithEdit.map((codeAction, i) => (
                        <div key={i} className="mb-2">
                            <label
                                className={`${buttonClassName} ${
                                    codeAction === activeCodeAction ? activeButtonClassName : inactiveButtonClassName
                                }`}
                                style={{ cursor: 'pointer' }}
                            >
                                <input
                                    type="radio"
                                    className="mr-2"
                                    checked={codeAction === activeCodeAction}
                                    onChange={e => {
                                        if (e.currentTarget.checked) {
                                            onCodeActionSetActive(codeAction)
                                        }
                                    }}
                                />
                                {codeAction.title}
                            </label>
                            {codeAction === activeCodeAction && (
                                <button
                                    className={`${buttonClassName} ${inactiveButtonClassName} btn-sm text-muted`}
                                    onClick={() => onCodeActionSetActive(undefined)}
                                >
                                    <CloseIcon className="icon-inline" /> Clear
                                </button>
                            )}
                        </div>
                    ))}
                </div>
            )}
            {codeActionsWithCommand.length > 0 && (
                <ul className="list-unstyled mb-0 d-flex flex-column flex-wrap">
                    {codeActionsWithCommand.map((codeAction, i) => (
                        <li key={i}>
                            <button
                                type="button"
                                className={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
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
                    buttonClassName={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
                />
            )}
        </div>
    )
}
