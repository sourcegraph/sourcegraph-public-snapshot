import { NotificationType } from '@sourcegraph/extension-api-classes'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { Redirect } from 'react-router'
import { Link } from 'react-router-dom'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ChangesetCreationStatus, createChangesetFromCodeAction } from '../changesets/preview/backend'
import { CreateChangesetFromCodeActionButton } from '../tasks/list/item/CreateChangesetFromCodeActionButton'
import { WorkspaceEditPreview } from '../threads/detail/inbox/item/WorkspaceEditPreview'
import { ActionsFormControl } from './internal/ActionsFormControl'

interface RenderChildrenProps {
    actions: React.ReactFragment
    preview?: React.ReactFragment
}

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps {
    actionsOrError: typeof LOADING | readonly sourcegraph.CodeAction[] | ErrorLike

    defaultPreview?: React.ReactFragment

    children: (props: RenderChildrenProps) => JSX.Element | null
}

/**
 * A form control to select actions and preview what will happen, or to invoke actions immediately.
 */
export const ActionsWithPreview: React.FunctionComponent<Props> = ({
    actionsOrError,
    extensionsController,
    defaultPreview,
    children,
    ...props
}) => {
    const onActionClick = useCallback(
        async (action: sourcegraph.CodeAction) => {
            try {
                if (action.command) {
                    await extensionsController.executeCommand(action.command)
                    if (action.diagnostics) {
                        // const fixedThisDiagnostic = action.diagnostics.some(
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

    const [activeAction, setActiveAction] = useState<sourcegraph.CodeAction | undefined>()

    const [createdThreadOrLoading, setCreatedThreadOrLoading] = useState<
        typeof LOADING | Pick<GQL.IDiscussionThread, 'idWithoutKind' | 'url' | 'status'>
    >()
    const [justCreated, setJustCreated] = useState(false)
    const onCreateThreadClick = useCallback(
        async (creationStatus: ChangesetCreationStatus) => {
            setCreatedThreadOrLoading(LOADING)
            try {
                const action = activeAction
                if (!action) {
                    throw new Error('no active code action')
                }
                setCreatedThreadOrLoading(
                    await createChangesetFromCodeAction({ extensionsController }, null, action, {
                        status: creationStatus,
                    })
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
        [activeAction, extensionsController]
    )

    return children({
        actions:
            actionsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(actionsOrError) ? (
                <span className="text-danger">{actionsOrError.message}</span>
            ) : (
                <ActionsFormControl
                    actions={actionsOrError}
                    activeAction={activeAction}
                    onActionClick={onActionClick}
                    onActionSetActive={setActiveAction}
                    className="py-2 px-3"
                    buttonClassName="btn py-0 px-2 text-decoration-none text-left"
                    inactiveButtonClassName="btn-link"
                    activeButtonClassName="border"
                />
            ),
        preview:
            activeAction && activeAction.edit ? (
                <>
                    <WorkspaceEditPreview
                        key={JSON.stringify(activeAction.edit)}
                        {...props}
                        extensionsController={extensionsController}
                        workspaceEdit={activeAction.edit}
                        className="overflow-auto p-2 mb-3"
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
                defaultPreview
            ),
    })
}
