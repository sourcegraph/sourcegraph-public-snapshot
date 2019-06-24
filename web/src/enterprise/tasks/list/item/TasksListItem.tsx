import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { Link, Redirect } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { fromDiagnostic, toDiagnostic } from '../../../../../../shared/src/api/extension/api/types'
import { CodeExcerpt } from '../../../../../../shared/src/components/CodeExcerpt'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { DiagnosticSeverityIcon } from '../../../../diagnostics/components/DiagnosticSeverityIcon'
import { fetchHighlightedFileLines } from '../../../../repo/backend'
import { ChangesetCreationStatus, createChangeset } from '../../../changesets/preview/backend'
import { DiagnosticInfo, getCodeActions } from '../../../threads/detail/backend'
import { WorkspaceEditPreview } from '../../../threads/detail/inbox/item/WorkspaceEditPreview'
import { CreateChangesetFromCodeActionButton } from './CreateChangesetFromCodeActionButton'
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
    // Reduce recomputation of code actions when the diagnostic object reference changes but it
    // contains the same data.
    const diagnosticData = JSON.stringify(fromDiagnostic(diagnostic))
    useEffect(() => {
        const diagnostic2: DiagnosticInfo = { ...toDiagnostic(JSON.parse(diagnosticData)), entry: diagnostic.entry }
        const subscriptions = new Subscription()
        subscriptions.add(
            getCodeActions({ diagnostic: diagnostic2, extensionsController })
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setCodeActionsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [diagnosticData, diagnostic.entry, extensionsController])

    const onCodeActionClick = useCallback(
        async (codeAction: sourcegraph.CodeAction) => {
            try {
                if (codeAction.command) {
                    await extensionsController.executeCommand(codeAction.command)
                    if (codeAction.diagnostics) {
                        // const fixedThisDiagnostic = codeAction.diagnostics.some(
                        //     d =>
                        //         d.code === diagnostic.code &&
                        //         d.message === diagnostic.message &&
                        //         d.source === diagnostic.source &&
                        //         d.severity === diagnostic.severity &&
                        //         d.range.isEqual(diagnostic.range)
                        // )
                        // TODO!(sqs)
                    }
                }
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error running action: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController]
    )

    const [activeCodeAction, setActiveCodeAction] = useState<sourcegraph.CodeAction | undefined>()

    const [createdThreadOrLoading, setCreatedThreadOrLoading] = useState<
        typeof LOADING | Pick<GQL.IDiscussionThread, 'idWithoutKind' | 'url' | 'status'>
    >()
    const [justCreated, setJustCreated] = useState(false)
    const onCreateThreadClick = useCallback(
        async (creationStatus: ChangesetCreationStatus) => {
            setCreatedThreadOrLoading(LOADING)
            try {
                const codeAction = activeCodeAction
                if (!codeAction) {
                    throw new Error('no active code action')
                }
                setCreatedThreadOrLoading(
                    await createChangeset({ extensionsController }, diagnostic, codeAction, creationStatus)
                )
                setJustCreated(true)
                setTimeout(() => setJustCreated(false), 2500)
            } catch (err) {
                setCreatedThreadOrLoading(undefined)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error creating changeset: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [activeCodeAction, diagnostic, extensionsController]
    )

    return (
        <div className={`d-flex flex-wrap align-items-stretch ${className}`}>
            <div style={{ flex: '1 1 40%', minWidth: '400px', maxWidth: '600px' }} className="pr-5">
                <header className={`d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                    <div className={`flex-1 d-flex align-items-center`}>
                        <h3 className="mb-0 small">
                            <LinkOrSpan to={diagnostic.entry.url} className="d-block">
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
                        buttonClassName="btn py-0 px-2 text-decoration-none text-left"
                        inactiveButtonClassName="btn-link"
                        activeButtonClassName="border"
                    />
                )}
            </div>
            <aside
                className="d-flex flex-column justify-content-between"
                style={{
                    flex: '2 0 60%',
                    minWidth: '600px',
                    margin: '-0.5rem -1rem -0.5rem 0',
                }}
            >
                {activeCodeAction && activeCodeAction.edit ? (
                    <>
                        <WorkspaceEditPreview
                            key={JSON.stringify(activeCodeAction.edit)}
                            {...props}
                            workspaceEdit={activeCodeAction.edit}
                            extensionsController={extensionsController}
                            className="tasks-list-item__workspace-edit-preview overflow-auto p-2 mb-3"
                        />
                        <div className="m-3">
                            {createdThreadOrLoading === undefined || createdThreadOrLoading === LOADING ? (
                                <CreateChangesetFromCodeActionButton
                                    onClick={onCreateThreadClick}
                                    isLoading={createdThreadOrLoading === LOADING}
                                />
                            ) : createdThreadOrLoading.status === GQL.ThreadStatus.PREVIEW ? (
                                <Redirect to={createdThreadOrLoading.url} push={true} />
                            ) : (
                                <>
                                    <Link className="btn btn-secondary" to={createdThreadOrLoading.url}>
                                        Changeset #{createdThreadOrLoading.idWithoutKind}
                                    </Link>
                                    {justCreated && <strong className="text-success ml-3">Created!</strong>}
                                </>
                            )}
                        </div>
                    </>
                ) : (
                    <CodeExcerpt
                        repoName={diagnostic.entry.repository.name}
                        commitID={diagnostic.entry.commit.oid}
                        filePath={diagnostic.entry.path}
                        context={4}
                        highlightRanges={[diagnostic.range].map(r => ({
                            line: r.start.line,
                            character: r.start.character,
                            highlightLength: r.end.character - r.start.character,
                        }))}
                        className="w-100 h-100 overflow-auto p-2"
                        isLightTheme={isLightTheme}
                        fetchHighlightedFileLines={fetchHighlightedFileLines}
                    />
                )}
            </aside>
        </div>
    )
}
