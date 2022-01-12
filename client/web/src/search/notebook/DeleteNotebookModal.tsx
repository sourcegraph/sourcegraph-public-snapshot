import Dialog from '@reach/dialog'
import React, { useCallback } from 'react'
import { useHistory } from 'react-router'
import { Observable } from 'rxjs'
import { mergeMap, startWith, tap, catchError } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { deleteNotebook as _deleteNotebook } from './backend'

interface DeleteNotebookProps {
    notebookId: string
    isOpen: boolean
    toggleDeleteModal: () => void
    deleteNotebook: typeof _deleteNotebook
}

const LOADING = 'loading' as const

export const DeleteNotebookModal: React.FunctionComponent<DeleteNotebookProps> = ({
    notebookId,
    deleteNotebook,
    isOpen,
    toggleDeleteModal,
}) => {
    const deleteLabelId = 'deleteNotebookId'
    const history = useHistory()

    const [onDelete, deleteCompletedOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
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
            [deleteNotebook, history, notebookId]
        )
    )

    return (
        <Dialog
            isOpen={isOpen}
            className="modal-body modal-body--centered p-4 rounded border"
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
        >
            <h3 className="text-danger" id={deleteLabelId}>
                Delete the notebook?
            </h3>

            <p>
                <strong>This action cannot be undone.</strong>
            </p>
            {(!deleteCompletedOrError || isErrorLike(deleteCompletedOrError)) && (
                <div className="text-right">
                    <button type="button" className="btn btn-outline-secondary mr-2" onClick={toggleDeleteModal}>
                        Cancel
                    </button>
                    <button type="button" className="btn btn-danger" onClick={onDelete}>
                        Yes, delete the notebook
                    </button>
                    {isErrorLike(deleteCompletedOrError) && (
                        <div className="alert alert-danger mt-2">
                            Error deleting notebook: {deleteCompletedOrError.message}
                        </div>
                    )}
                </div>
            )}
            {deleteCompletedOrError && <div>{deleteCompletedOrError === 'loading' && <LoadingSpinner />}</div>}
        </Dialog>
    )
}
