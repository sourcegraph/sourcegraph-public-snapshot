import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useEffect, useState, useCallback } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { LinkOrSpan } from '../../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { DiagnosticSeverityIcon } from '../../../../../diagnostics/components/DiagnosticSeverityIcon'
import { ThreadSettings } from '../../../settings'
import { DiagnosticInfo, getCodeActions, diagnosticID, codeActionID, getActiveCodeAction0 } from '../../backend'
import { ThreadInboxItemActions } from './actions/ThreadInboxItemActions'
import { WorkspaceEditPreview } from './WorkspaceEditPreview'
import { updateThreadSettings } from '../../../../../discussions/backend'
import FileOutlineIcon from 'mdi-react/FileOutlineIcon'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    diagnostic: DiagnosticInfo
    className?: string
    headerClassName?: string
    headerStyle?: React.CSSProperties
    isLightTheme: boolean
    history: H.History
    location: H.Location
}

/**
 * An inbox item in a thread that refers to a file.
 */
export const ThreadInboxFileItem: React.FunctionComponent<Props> = ({
    thread,
    threadSettings,
    onThreadUpdate,
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

    const onCodeActionActivate = useCallback(
        async (codeAction: sourcegraph.CodeAction | undefined) => {
            onThreadUpdate(
                await updateThreadSettings(thread, {
                    ...threadSettings,
                    actions: {
                        ...threadSettings.actions,
                        [diagnosticID(diagnostic)]: codeAction ? codeActionID(codeAction) : undefined,
                    },
                })
            )
        },
        [diagnostic, onThreadUpdate, thread, threadSettings]
    )

    const activeCodeAction =
        codeActionsOrError !== LOADING && !isErrorLike(codeActionsOrError)
            ? getActiveCodeAction0(diagnostic, threadSettings, codeActionsOrError)
            : undefined

    return (
        <div className={`card border ${className}`}>
            <header className={`card-header d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                <div className="flex-1 d-flex align-items-center">
                    <h3 className="mb-0 h6">
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
            <div className="d-flex align-items-center mt-2 mx-2 mb-1">
                <DiagnosticSeverityIcon severity={diagnostic.severity} className="icon-inline mr-1" />
                <span>{diagnostic.message}</span>
            </div>
            {codeActionsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(codeActionsOrError) ? (
                <span className="text-danger">{codeActionsOrError.message}</span>
            ) : (
                <>
                    <ThreadInboxItemActions
                        {...props}
                        thread={thread}
                        onThreadUpdate={onThreadUpdate}
                        threadSettings={threadSettings}
                        codeActions={codeActionsOrError}
                        activeCodeAction={activeCodeAction}
                        onCodeActionActivate={onCodeActionActivate}
                        className="px-2 pt-2 pb-0"
                        buttonClassName="btn px-1 py-0 text-decoration-none"
                        inactiveButtonClassName="btn-link"
                        activeButtonClassName="btn-primary"
                        extensionsController={extensionsController}
                    />
                    {activeCodeAction ? (
                        activeCodeAction.edit ? (
                            <WorkspaceEditPreview
                                key={JSON.stringify(activeCodeAction.edit)}
                                {...props}
                                workspaceEdit={activeCodeAction.edit}
                                extensionsController={extensionsController}
                                className="border-top overflow-auto"
                            />
                        ) : (
                            'no edit'
                        )
                    ) : (
                        'no active code action'
                    )}
                </>
            )}
        </div>
    )
}
