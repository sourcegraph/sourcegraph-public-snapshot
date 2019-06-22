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
import { DiagnosticInfo, getCodeActions } from '../../../threads/detail/backend'
import { WorkspaceEditPreview } from '../../../threads/detail/inbox/item/WorkspaceEditPreview'
import { TasksListItemActions } from './TasksListItemActions'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
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
    thread,
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

    const onCodeActionClick = useCallback((codeAction: sourcegraph.CodeAction) => {
        // TODO!(sqs)
        alert('not implemented')
    }, [])

    return (
        <div className={className}>
            <header className={`d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                <div className={`flex-1 d-flex align-items-center`}>
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
            <div className={`d-flex align-items-center mt-2 mb-1`}>
                <DiagnosticSeverityIcon severity={diagnostic.severity} className="icon-inline mr-1" />
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
                    onCodeActionClick={onCodeActionClick}
                    className="pt-2 pb-0"
                    buttonClassName="btn btn-link px-2 py-0 text-decoration-none"
                />
            )}
        </div>
    )
}
