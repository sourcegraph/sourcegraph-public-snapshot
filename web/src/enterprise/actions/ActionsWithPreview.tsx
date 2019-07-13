import { Diagnostic } from '@sourcegraph/extension-api-types'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { Action } from '../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { useEffectAsync } from '../../util/useEffectAsync'
import { computeDiff, FileDiff } from '../threads/detail/changes/computeDiff'
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
    diagnostic: Diagnostic // TODO!(sqs): this is a weird api

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
    diagnostic,
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

    const [fileDiffsOrError, setFileDiffsOrError] = useState<typeof LOADING | null | FileDiff[] | ErrorLike>(LOADING)
    useEffectAsync(async () => {
        setFileDiffsOrError(LOADING) // TODO!(sqs) causes jitter
        try {
            setFileDiffsOrError(
                selectedAction && selectedAction.computeEdit
                    ? await computeDiff(extensionsController, [
                          { actionEditCommand: selectedAction.computeEdit, diagnostic },
                      ])
                    : null
            )
        } catch (err) {
            setFileDiffsOrError(asError(err))
        }
    }, [actionsOrError, diagnostic, extensionsController, selectedAction])

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
            fileDiffsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(fileDiffsOrError) ? (
                <span className="text-danger">{fileDiffsOrError.message}</span>
            ) : fileDiffsOrError ? (
                <WorkspaceEditPreview {...props} fileDiffs={fileDiffsOrError} className="overflow-auto p-2 mb-3" />
            ) : (
                defaultPreview
            ),
    })
}
