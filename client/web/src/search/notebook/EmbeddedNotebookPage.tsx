import { noop } from 'lodash'
import React, { useEffect, useMemo } from 'react'
import { useLocation } from 'react-router-dom'
import { catchError, startWith } from 'rxjs/operators'

import { asError, isErrorLike, isMacPlatform } from '@sourcegraph/common'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { Alert, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { createPlatformContext } from '../../platform/context'
import { fetchHighlightedFileLineRanges, fetchRepository, resolveRevision } from '../../repo/backend'
import { eventLogger } from '../../tracking/eventLogger'

import { fetchNotebook } from './backend'
import { NotebookContent, NotebookContentProps } from './NotebookContent'

interface EmbeddedNotebookPageProps
    extends Pick<
        NotebookContentProps,
        | 'isLightTheme'
        | 'searchContextsEnabled'
        | 'showSearchContext'
        | 'isSourcegraphDotCom'
        | 'authenticatedUser'
        | 'settingsCascade'
    > {
    notebookId: string
}

const LOADING = 'loading' as const

export const EmbeddedNotebookPage: React.FunctionComponent<EmbeddedNotebookPageProps> = ({ notebookId, ...props }) => {
    useEffect(() => eventLogger.logViewEvent('EmbeddedNotebookPage'), [])

    const platformContext = useMemo(() => createPlatformContext(), [])
    const extensionsController = useMemo(() => createExtensionsController(platformContext), [platformContext])
    const isMacPlatformMemoized = useMemo(() => isMacPlatform(), [])

    const notebookOrError = useObservable(
        useMemo(
            () =>
                fetchNotebook(notebookId).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [notebookId]
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
                    globbing={true}
                    isMacPlatform={isMacPlatformMemoized}
                    fetchRepository={fetchRepository}
                    fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                    resolveRevision={resolveRevision}
                    streamSearch={aggregateStreamingSearch}
                    telemetryService={eventLogger}
                    platformContext={platformContext}
                    extensionsController={extensionsController}
                />
            )}
        </div>
    )
}
