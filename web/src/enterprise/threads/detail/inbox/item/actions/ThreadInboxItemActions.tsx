import H from 'history'
import CheckboxBlankCirckeOutlineIcon from 'mdi-react/CheckboxBlankCircleOutlineIcon'
import CheckboxMarkedCircleOutlineIcon from 'mdi-react/CheckboxMarkedCircleOutlineIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import React, { useCallback, useState } from 'react'
import { CodeAction } from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../../shared/src/graphql/schema'
import { ThreadSettings } from '../../../../settings'
import { ThreadInboxItemActionsDropdownButton } from './ThreadInboxItemActionsDropdownButton'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    codeActions: CodeAction[]
    activeCodeAction: CodeAction | undefined
    onCodeActionActivate: (codeAction: CodeAction | undefined) => void

    className?: string
    buttonClassName?: string
    inactiveButtonClassName?: string
    activeButtonClassName?: string
    history: H.History
    location: H.Location
}

/**
 * The actions that can be performed on an item in a thread inbox.
 *
 * TODO!(sqs): dedupe with TasksListItemActions?
 */
// tslint:disable: jsx-no-lambda
export const ThreadInboxItemActions: React.FunctionComponent<Props> = ({
    codeActions,
    activeCodeAction,
    onCodeActionActivate: onCodeActionClick,
    className,
    buttonClassName = 'btn btn-link text-decoration-none',
    inactiveButtonClassName,
    activeButtonClassName,
}) => {
    const a = 123
    const codeActionsWithEdit = codeActions.filter(({ edit }) => !!edit)
    const codeActionsWithCommand = codeActions.filter(
        ({ diagnostics, command }) => !!command && diagnostics && diagnostics.length > 0
    )
    const codeActionsWithCommandSecondary = codeActions.filter(
        ({ diagnostics, command }) => !!command && (!diagnostics || diagnostics.length === 0)
    )

    return (
        <div className={className}>
            {(codeActionsWithEdit.length > 0 ||
                codeActionsWithCommand.length > 0 ||
                codeActionsWithCommandSecondary.length > 0) && (
                <div className={`d-flex align-items-center`}>
                    {codeActionsWithEdit.length > 0 && (
                        <div className="btn-group btn-group-toggle" data-toggle="buttons">
                            {codeActionsWithEdit.map((codeAction, i) => (
                                <label
                                    key={i}
                                    className={`d-flex align-items-center ${buttonClassName} ${
                                        codeAction === activeCodeAction
                                            ? activeButtonClassName
                                            : inactiveButtonClassName
                                    } mr-2 mb-2 rounded`}
                                    style={{ cursor: 'pointer' }}
                                    onClick={() => onCodeActionClick(codeAction)}
                                >
                                    <input
                                        type="radio"
                                        className="mr-2"
                                        autoComplete="off"
                                        checked={codeAction === activeCodeAction}
                                        onChange={e =>
                                            onCodeActionClick(e.currentTarget.checked ? codeAction : undefined)
                                        }
                                    />{' '}
                                    {/* TODO!(sqs) {codeAction === activeCodeAction ? (
                                        <CheckboxMarkedCircleOutlineIcon className="icon-inline small mr-1" />
                                    ) : (
                                        <CheckboxBlankCirckeOutlineIcon className="icon-inline small mr-1" />
                                    )} */}{' '}
                                    {codeAction.title}
                                </label>
                            ))}
                        </div>
                    )}
                    {codeActionsWithCommand.length > 0 && (
                        <ul className="list-unstyled mb-0 d-flex flex-wrap">
                            {codeActionsWithCommand.map((codeAction, i) => (
                                <button
                                    key={i}
                                    type="button"
                                    className={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
                                    onClick={() => alert('TODO!(sqs)')}
                                >
                                    {codeAction.title}
                                </button>
                            ))}
                        </ul>
                    )}
                    {codeActionsWithCommandSecondary.length > 0 && (
                        <ThreadInboxItemActionsDropdownButton
                            codeActions={codeActionsWithCommandSecondary}
                            buttonClassName={`${buttonClassName} ${inactiveButtonClassName} mr-2 mb-2`}
                        />
                    )}
                </div>
            )}
        </div>
    )
}
