import { noop } from 'lodash'
import React, { useEffect, useMemo } from 'react'
import { useLocation } from 'react-router-dom'
import { catchError, startWith } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Alert, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { fetchNotebook as _fetchNotebook } from './backend'
import { NotebookContent, NotebookContentProps } from './NotebookContent'

interface EmbeddedNotebookPageProps
    extends Omit<NotebookContentProps, 'blocks' | 'onUpdateBlocks' | 'location' | 'viewerCanManage'> {
    notebookId: string
    fetchNotebook?: typeof _fetchNotebook
}

const LOADING = 'loading' as const

export const EmbeddedNotebookPage: React.FunctionComponent<EmbeddedNotebookPageProps> = ({
    notebookId,
    fetchNotebook = _fetchNotebook,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent('EmbeddedNotebookPage'), [props.telemetryService])

    const notebookOrError = useObservable(
        useMemo(
            () =>
                fetchNotebook(notebookId).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [fetchNotebook, notebookId]
        )
    )

    const location = useLocation()
    return (
        <div className="p-3">
            {notebookOrError === LOADING && (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner />
                </div>
            )}
            {isErrorLike(notebookOrError) && (
                <Alert variant="danger">
                    Error while loading the notebook: <strong>{notebookOrError.message}</strong>
                </Alert>
            )}
            {notebookOrError && notebookOrError !== LOADING && !isErrorLike(notebookOrError) && (
                <NotebookContent
                    {...props}
                    location={location}
                    blocks={notebookOrError.blocks}
                    onUpdateBlocks={noop}
                    viewerCanManage={false}
                />
            )}
        </div>
    )
}
