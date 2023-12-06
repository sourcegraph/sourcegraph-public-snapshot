import React, { type FC, useCallback, useEffect } from 'react'

import { useNavigate } from 'react-router-dom'
import type { Observable } from 'rxjs'
import { mergeMap, startWith, tap, catchError } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useEventObservable, Modal, Button, Alert, H3, Text } from '@sourcegraph/wildcard'

import type { deleteNotebook as _deleteNotebook } from '../backend'

interface DeleteNotebookModalProps extends TelemetryProps, TelemetryV2Props {
    notebookId: string
    isOpen: boolean
    toggleDeleteModal: () => void
    deleteNotebook: typeof _deleteNotebook
}

const LOADING = 'loading' as const
const deleteLabelId = 'deleteNotebookId'

export const DeleteNotebookModal: FC<DeleteNotebookModalProps> = ({
    notebookId,
    deleteNotebook,
    isOpen,
    toggleDeleteModal,
    telemetryService,
    telemetryRecorder,
}) => {
    const navigate = useNavigate()

    useEffect(() => {
        if (isOpen) {
            telemetryService.log('SearchNotebookDeleteModalOpened')
            telemetryRecorder.recordEvent('SearchNotebookDeleteModal', 'opened')
        }
    }, [isOpen, telemetryService, telemetryRecorder])

    const [onDelete, deleteCompletedOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    tap(() => {
                        telemetryService.log('SearchNotebookDeleteButtonClicked')
                        telemetryRecorder.recordEvent('SearchNotebookDeleteButton', 'clicked')
                    }),
                    mergeMap(() =>
                        deleteNotebook(notebookId).pipe(
                            tap(() => {
                                navigate('/notebooks')
                            }),
                            startWith(LOADING),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [deleteNotebook, navigate, notebookId, telemetryService, telemetryRecorder]
        )
    )

    return (
        <Modal isOpen={isOpen} position="center" onDismiss={toggleDeleteModal} aria-labelledby={deleteLabelId}>
            <H3 className="text-danger" id={deleteLabelId}>
                Delete the notebook?
            </H3>

            <Text>
                <strong>This action cannot be undone.</strong>
            </Text>
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
