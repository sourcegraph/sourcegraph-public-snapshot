import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { DiagnosticSeverityIcon } from '../../../../diagnostics/components/DiagnosticSeverityIcon'
import { updateThreadSettings } from '../../../../discussions/backend'
import { createPreviewChangeset } from '../../../changesets/preview/backend'
import {
    codeActionID,
    diagnosticID,
    DiagnosticInfo,
    getActiveCodeAction0,
    getCodeActions,
} from '../../../threads/detail/backend'
import { WorkspaceEditPreview } from '../../../threads/detail/inbox/item/WorkspaceEditPreview'
import { ThreadSettings } from '../../../threads/settings'
import { TasksListItemActions } from './TasksListItemActions'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    diagnostic: DiagnosticInfo

    className?: string
    headerClassName?: string
    headerStyle?: React.CSSProperties
    isLightTheme: boolean
    history: H.History
    location: H.Location
}

/**
 * An item in a task list.
 */
export const TasksListItem: React.FunctionComponent<Props> = ({
    diagnostic,
    className = '',
    headerClassName = '',
    headerStyle,
    isLightTheme,
    extensionsController,
    ...props
}) => {
    const [codeActionsOrError, setCodeActionsOrError] = useState<typeof LOADING | sourcegraph.CodeAction[] | ErrorLike>(
        LOADING
    )
    // tslint:disable-next-line: no-floating-promises
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            getCodeActions({ diagnostic, extensionsController })
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setCodeActionsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [diagnostic, extensionsController])

    const onCodeActionClick = useCallback(
        async (codeAction: sourcegraph.CodeAction) => {
            if (codeAction.command) {
                await extensionsController.executeCommand(codeAction.command)
            }
            if (codeAction.edit) {
                // TODO!(sqs): show loading
                const changeset = await createPreviewChangeset({ extensionsController }, codeAction)
                props.history.push(changeset.url)
            }
        },
        [extensionsController, props.history]
    )

    const [activeCodeAction, setActiveCodeAction] = useState<sourcegraph.CodeAction | undefined>()

    return (
        <div className={`d-flex flex-wrap ${className}`}>
            <div style={{ flex: '1 1 33%' }} className="pr-5">
                <header className={`d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                    <div className={`flex-1 d-flex align-items-center`}>
                        <h3 className="mb-0 small">
                            <LinkOrSpan to={diagnostic.entry.url || 'TODO!(sqs)'} className="d-block">
                                {diagnostic.entry.path ? (
                                    <>
                                        <span className="font-weight-normal">
                                            {displayRepoName(diagnostic.entry.repository.name)}
                                        </span>{' '}
                                        › {diagnostic.entry.path}
                                    </>
                                ) : (
                                    displayRepoName(diagnostic.entry.repository.name)
                                )}
                            </LinkOrSpan>
                        </h3>
                    </div>
                </header>
                <div className={`d-flex align-items-start mt-2 mb-1`}>
                    <DiagnosticSeverityIcon severity={diagnostic.severity} className="icon-inline mr-2" />
                    <span>{diagnostic.message}</span>
                </div>
                {codeActionsOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(codeActionsOrError) ? (
                    <span className="text-danger">{codeActionsOrError.message}</span>
                ) : (
                    <TasksListItemActions
                        {...props}
                        codeActions={codeActionsOrError}
                        activeCodeAction={activeCodeAction}
                        onCodeActionClick={onCodeActionClick}
                        onCodeActionSetActive={setActiveCodeAction}
                        className="pt-2 pb-0"
                        buttonClassName="btn py-0 px-2 text-decoration-none"
                        inactiveButtonClassName="btn-link"
                        activeButtonClassName="border"
                    />
                )}
            </div>
            <aside style={{ flex: '2 0 66%', minWidth: '800px' }}>
                {activeCodeAction && activeCodeAction.edit && (
                    <WorkspaceEditPreview
                        key={JSON.stringify(activeCodeAction.edit)}
                        {...props}
                        workspaceEdit={activeCodeAction.edit}
                        extensionsController={extensionsController}
                        className="overflow-auto"
                    />
                )}
            </aside>
        </div>
    )
}
