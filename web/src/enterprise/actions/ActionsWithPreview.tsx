import { NotificationType } from '@sourcegraph/extension-api-classes'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { Action } from '../../../../shared/src/api/types/action'
import { WorkspaceEdit } from '../../../../shared/src/api/types/workspaceEdit'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ChangesetCreationStatus, createChangesetFromCodeAction } from '../changesets/preview/backend'
import { ChangesetButtonOrLinkExistingChangeset, PENDING_CREATION } from '../tasks/list/item/ChangesetButtonOrLink'
import { WorkspaceEditPreview } from '../threads/detail/inbox/item/WorkspaceEditPreview'
import { ActionsFormControl } from './internal/ActionsFormControl'

interface RenderChildrenProps {
    actions: React.ReactFragment
    preview?: React.ReactFragment
}

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerProps {
    actionsOrError: typeof LOADING | readonly Action[] | ErrorLike
    selectedAction: Action | null
    onActionSelect: (action: Action | null) => void

    defaultPreview?: React.ReactFragment

    children: (props: RenderChildrenProps) => JSX.Element | null
}

/**
 * A form control to select actions and preview what will happen, or to invoke actions immediately.
 */
export const ActionsWithPreview: React.FunctionComponent<Props> = ({
    actionsOrError,
    selectedAction,
    onActionSelect,
    extensionsController,
    defaultPreview,
    children,
    ...props
}) => {
    const onActionSetSelected = useCallback(
        (value: boolean, action: Action): void => {
            if (value) {
                onActionSelect(action)
            } else if (action === selectedAction) {
                onActionSelect(null)
            }
        },
        [onActionSelect, selectedAction]
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
                    selectedAction={selectedAction}
                    onActionSetSelected={onActionSetSelected}
                    className="mt-4"
                    buttonClassName="btn px-2 py-1"
                    activeButtonClassName="btn-primary"
                    inactiveButtonClassName="btn-link"
                />
            ),
        preview:
            selectedAction && selectedAction.edit ? (
                <WorkspaceEditPreview
                    key={JSON.stringify(selectedAction.edit)}
                    {...props}
                    extensionsController={extensionsController}
                    workspaceEdit={WorkspaceEdit.fromJSON(selectedAction.edit)}
                    className="overflow-auto p-2 mb-3"
                />
            ) : (
                defaultPreview
            ),
    })
}
