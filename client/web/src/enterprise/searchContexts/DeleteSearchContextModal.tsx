import React, { useCallback } from 'react'

import { useHistory } from 'react-router'
import { Observable } from 'rxjs'
import { mergeMap, startWith, tap, catchError } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ISearchContext } from '@sourcegraph/shared/src/schema'
import { Button, LoadingSpinner, useEventObservable, Modal, Alert, Typography } from '@sourcegraph/wildcard'

import { ALLOW_NAVIGATION } from '../../components/AwayPrompt'

interface DeleteSearchContextModalProps
    extends Pick<SearchContextProps, 'deleteSearchContext'>,
        PlatformContextProps<'requestGraphQL'> {
    isOpen: boolean
    searchContext: ISearchContext
    toggleDeleteModal: () => void
}

export const DeleteSearchContextModal: React.FunctionComponent<
    React.PropsWithChildren<DeleteSearchContextModalProps>
> = ({ isOpen, deleteSearchContext, toggleDeleteModal, searchContext, platformContext }) => {
    const LOADING = 'loading' as const
    const deleteLabelId = 'deleteSearchContextId'
    const history = useHistory()

    const [onDelete, deleteCompletedOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                click.pipe(
                    mergeMap(() =>
                        deleteSearchContext(searchContext.id, platformContext).pipe(
                            tap(() => {
                                history.push('/contexts', ALLOW_NAVIGATION)
                            }),
                            startWith(LOADING),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [deleteSearchContext, history, searchContext, platformContext]
        )
    )

    return (
        <Modal
            position="center"
            isOpen={isOpen}
            onDismiss={toggleDeleteModal}
            aria-labelledby={deleteLabelId}
            data-testid="delete-search-context-modal"
        >
            <Typography.H3 className="text-danger" id={deleteLabelId}>
                Delete search context?
            </Typography.H3>

            <p>
                <strong>This action cannot be undone.</strong>
            </p>
            {(!deleteCompletedOrError || isErrorLike(deleteCompletedOrError)) && (
                <div className="text-right">
                    <Button className="mr-2" onClick={toggleDeleteModal} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <Button data-testid="confirm-delete-search-context" onClick={onDelete} variant="danger">
                        Yes, delete search context
                    </Button>
                    {isErrorLike(deleteCompletedOrError) && (
                        <Alert variant="danger">Error deleting search context: {deleteCompletedOrError.message}</Alert>
                    )}
                </div>
            )}
            {deleteCompletedOrError && <div>{deleteCompletedOrError === 'loading' && <LoadingSpinner />}</div>}
        </Modal>
    )
}
