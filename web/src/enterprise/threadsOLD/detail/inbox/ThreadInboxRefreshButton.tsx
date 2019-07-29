import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import ReloadIcon from 'mdi-react/ReloadIcon'
import React, { useCallback, useState } from 'react'
import { of, throwError } from 'rxjs'
import { first, switchMap } from 'rxjs/operators'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { addTargetToThread, fetchDiscussionThreadAndComments } from '../../../../discussions/backend'
import { search } from '../../../../search/backend'
import { ThreadSettings } from '../../settings'

const flattenSearchResults = (resultSets: GQL.ISearchResults[]): GQL.ISearchResults | null => {
    if (resultSets.length === 0) {
        return null
    }
    const flattened = resultSets[0]
    for (const r of resultSets.slice(1)) {
        flattened.results = [...flattened.results, ...r.results]
        flattened.resultCount += r.resultCount
        flattened.limitHit = flattened.limitHit || r.limitHit
    }
    return flattened
}

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    className?: string
    buttonClassName?: string
}

/**
 * A button that refreshes the contents of a thread's inbox.
 */
export const ThreadInboxRefreshButton: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    className = '',
    buttonClassName = 'btn-secondary',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const inboxItems: Pick<GQL.ISearchResults, 'results' | 'resultCount'> = flattenSearchResults(
                    await Promise.all(
                        (threadSettings.queries || []).map(query =>
                            search(query, { extensionsController })
                                .pipe(
                                    first(),
                                    switchMap(r => {
                                        if (isErrorLike(r)) {
                                            // tslint:disable-next-line: rxjs-throw-error
                                            return throwError(r)
                                        }
                                        return of(r)
                                    })
                                )
                                .toPromise()
                        )
                    )
                ) || { results: [], resultCount: 0 }

                await Promise.all(
                    inboxItems.results
                        .filter((r): r is GQL.IFileMatch => r.__typename === 'FileMatch')
                        .map(async item =>
                            addTargetToThread({
                                threadID: thread.id,
                                target: {
                                    repo: {
                                        repositoryID: item.repository.id,
                                        revision: item.file.commit.oid,
                                        path: item.file.path,
                                        selection: {
                                            startLine: item.lineMatches[0].lineNumber,
                                            endLine: item.lineMatches[0].lineNumber,
                                            startCharacter: item.lineMatches[0].offsetAndLengths[0][0],
                                            endCharacter:
                                                item.lineMatches[0].offsetAndLengths[0][0] +
                                                item.lineMatches[0].offsetAndLengths[0][1],
                                        },
                                    },
                                },
                            }).toPromise()
                        )
                )

                onThreadUpdate(await fetchDiscussionThreadAndComments(thread.idWithoutKind).toPromise())
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error refreshing: ${err.message}`,
                    type: NotificationType.Error,
                })
            } finally {
                setIsLoading(false)
            }
        },
        [threadSettings.queries, onThreadUpdate, thread.idWithoutKind, thread.id, extensionsController]
    )
    return (
        <button type="submit" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <ReloadIcon className="icon-inline" />}
        </button>
    )
}
