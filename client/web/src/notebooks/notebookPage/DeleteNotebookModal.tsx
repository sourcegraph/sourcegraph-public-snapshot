import React, { useCallback, useEffect } from 'react'

import { useHistory } from 'react-router'
import { Observable } from 'rxjs'
import { mergeMap, startWith, tap, catchError } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useEventObservable, Modal, Button, Alert, Typography } from '@sourcegraph/wildcard'

import { deleteNotebook as _deleteNotebook } from '../backend'

interface DeleteNotebookModalProps extends TelemetryProps {
    notebookId: string
    isOpen: boolean
    toggleDeleteModal: () => void
    deleteNotebook: typeof _deleteNotebook
}

const LOADING = 'loading' as const

export const DeleteNotebookModal: React.FunctionComponent<React.PropsWithChildren<DeleteNotebookModalProps>> = ({
    notebookId,
    deleteNotebook,
    isOpen,
    toggleDeleteModal,
    telemetryService,
}) => {
    useEffect(() => {
        if (isOpen) {
            telemetryService.log('SearchNotebookDeleteModalOpened')
        }
    }, [isOpen, telemetryService])
    const deleteLabelId = 'deleteNotebookId'
    const history = useHistory()

    const [onDelete, deleteCompletedOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    tap(() => telemetryService.log('SearchNotebookDeleteButtonClicked')),
                    mergeMap(() =>
                        deleteNotebook(notebookId).pipe(
                            tap(() => {
                                history.push('/notebooks')
                            }),
                            startWith(LOADING),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [deleteNotebook, history, notebookId, telemetryService]
        )
    )

    return (
        <Modal isOpen={isOpen} position="center" onDismiss={toggleDeleteModal} aria-labelledby={deleteLabelId}>
            <Typography.H3 className="text-danger" id={deleteLabelId}>
                Delete the notebook?
            </Typography.H3>

            <p>
                <strong>This action cannot be undone.</strong>
            </p>
            {(!deleteCompletedOrError || isErrorLike(deleteCompletedOrError)) && (
                <div className="text-right">
                    <Button className="mr-2" onClick={toggleDeleteModal} variant="secondary" outline={true}>
                        Cancel
                    </Button>
                    <Button onClick={onDelete} variant="danger">
                        Yes, delete the notebook
                    </Button>
                    {isErrorLike(deleteCompletedOrError) && (
                        <Alert className="mt-2" variant="danger">
                            Error deleting notebook: {deleteCompletedOrError.message}
                        </Alert>
                    )}
                </div>
            )}
            {deleteCompletedOrError && <div>{deleteCompletedOrError === 'loading' && <LoadingSpinner />}</div>}
        </Modal>
    )
}
