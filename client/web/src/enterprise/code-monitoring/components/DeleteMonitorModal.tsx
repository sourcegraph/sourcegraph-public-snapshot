import Dialog from '@reach/dialog'
import React, { useCallback } from 'react'
import { Observable, throwError } from 'rxjs'
import { mergeMap, startWith, tap, catchError } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { CodeMonitorFormProps } from './CodeMonitorForm'

interface DeleteModalProps extends Pick<CodeMonitorFormProps, 'history' | 'codeMonitor'> {
    isOpen: boolean
    toggleDeleteModal: () => void
    deleteCodeMonitor: (id: string) => Observable<void>
}

export const DeleteMonitorModal: React.FunctionComponent<DeleteModalProps> = ({
    history,
    isOpen,
    deleteCodeMonitor,
    toggleDeleteModal,
    codeMonitor,
}) => {
    const LOADING = 'loading' as const

    const deleteLabelId = 'deleteCodeMonitor'

    const [onDelete, deleteCompletedOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    mergeMap(() => {
                        if (codeMonitor) {
                            return deleteCodeMonitor(codeMonitor.id).pipe(
                                tap(() => {
                                    history.push('/code-monitoring')
                                }),
                                startWith(LOADING),
                                catchError(error => [asError(error)])
                            )
                        }

                        return throwError(new Error('Failed to delete: Code monitor ID not provided'))
                    })
                ),
            [deleteCodeMonitor, history, codeMonitor]
        )
    )

    return (
        <Dialog
            isOpen={isOpen}
            className="modal-body modal-body--centered p-4 rounded border"
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
            data-testid="delete-modal"
        >
            <h3 className="text-danger" id={deleteLabelId}>
                Delete code monitor?
            </h3>

            <p>
                <strong>This action cannot be undone.</strong> Code monitoring will no longer watch for trigger event
                and all actions will immediately be removed.
            </p>
            {(!deleteCompletedOrError || isErrorLike(deleteCompletedOrError)) && (
                <div className="text-right">
                    <Button className="mr-2" onClick={toggleDeleteModal} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <Button onClick={onDelete} data-testid="confirm-delete-monitor" variant="danger">
                        Yes, delete code monitor
                    </Button>
                    {isErrorLike(deleteCompletedOrError) && (
                        <div className="alert-danger">Error deleting monitor: {deleteCompletedOrError.message}</div>
                    )}
                </div>
            )}
            {deleteCompletedOrError && <div>{deleteCompletedOrError === 'loading' && <LoadingSpinner />}</div>}
        </Dialog>
    )
}
