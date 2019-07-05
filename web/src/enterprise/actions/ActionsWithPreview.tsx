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
import {
    ChangesetButtonOrLink,
    ChangesetButtonOrLinkExistingChangeset,
    PENDING_CREATION,
} from '../tasks/list/item/ChangesetButtonOrLink'
import { CreateOrPreviewChangesetButton } from '../tasks/list/item/CreateOrPreviewChangesetButton'
import { WorkspaceEditPreview } from '../threads/detail/inbox/item/WorkspaceEditPreview'
import { ActionsFormControl } from './internal/ActionsFormControl'
import { useOnActionClickCallback } from './useOnActionClickCallback'

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
    const onActionClick = useOnActionClickCallback(extensionsController)

    const [activeAction, setActiveAction] = useState<sourcegraph.CodeAction | undefined>()

    const [createdThreadOrLoading, setCreatedThreadOrLoading] = useState<ChangesetButtonOrLinkExistingChangeset>(
        LOADING
    )
    const [justCreated, setJustCreated] = useState(false)
    const onCreateThreadClick = useCallback(
        async (creationStatus: ChangesetCreationStatus) => {
            setCreatedThreadOrLoading(PENDING_CREATION)
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
                setCreatedThreadOrLoading(null)
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
                    className="mt-4"
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
                    <ChangesetButtonOrLink
                        onClick={onCreateThreadClick}
                        existingChangeset={createdThreadOrLoading}
                        className="m-3"
                    />
                </>
            ) : (
                defaultPreview
            ),
    })
}
