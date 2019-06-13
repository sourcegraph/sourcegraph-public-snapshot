import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CloseCircleIcon from 'mdi-react/CloseCircleIcon'
import DotsHorizontalCircleIcon from 'mdi-react/DotsHorizontalCircleIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React, { useEffect, useMemo, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { ListHeaderQueryLinksNav } from '../../../components/ListHeaderQueryLinks'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { PullRequest, ThreadSettings } from '../../../settings'
import { Changeset, computeChangesets, getChangesetExternalStatus } from '../../backend'
import { ThreadStatusItemsProgressBar } from '../ThreadStatusItemsProgressBar'
import { ThreadActionsPullRequestListHeaderFilterButtonDropdown } from './ThreadActionsPullRequestListHeaderFilterButtonDropdown'
import { ThreadActionsPullRequestsListItem } from './ThreadActionsPullRequestsListItem'

interface Props extends QueryParameterProps, ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'url'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    action?: React.ReactFragment

    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The list of pull requests associated with a thread.
 */
export const ThreadActionsPullRequestsList: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    query,
    onQueryChange,
    action,
    location,
    extensionsController,
}) => {
    const [changesetsOrError, setChangesetsOrError] = useState<typeof LOADING | Changeset[] | ErrorLike>(LOADING)
    // tslint:disable-next-line: no-floating-promises
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            computeChangesets(extensionsController, thread, threadSettings)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setChangesetsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [thread.id, threadSettings, extensionsController, query, thread])

    const itemsOrError: typeof LOADING | PullRequest[] | ErrorLike =
        changesetsOrError !== LOADING && !isErrorLike(changesetsOrError)
            ? changesetsOrError.map(c => ({
                  ...c.pullRequest,
                  ...getChangesetExternalStatus(c),
                  items: 'x'.repeat(c.fileDiffs.length).split('x'),
                  number: parseInt(getChangesetExternalStatus(c).title.replace('#', '')),
                  repo: c.repo,
                  updatedAt: '2019-05-30',
                  updatedBy: 'alice',
              }))
            : changesetsOrError

    const filteredItemsOrError =
        itemsOrError !== LOADING && !isErrorLike(itemsOrError)
            ? {
                  items: itemsOrError,
                  filteredItems: itemsOrError.filter(
                      item =>
                          // (query.includes('is:pending') && item.status === 'pending') ||
                          (query.includes('is:open') && item.status === 'open') ||
                          (query.includes('is:merged') && item.status === 'merged') ||
                          (query.includes('is:closed') && item.status === 'closed') ||
                          !query.includes('is:')
                  ),
              }
            : itemsOrError

    return (
        <div className="thread-actions-pull-requests-list">
            {isErrorLike(filteredItemsOrError) ? (
                <div className="alert alert-danger mt-2">{filteredItemsOrError.message}</div>
            ) : (
                <div className="card">
                    <div className="card-header d-flex align-items-center justify-content-between">
                        <div className="form-check mx-2">
                            <input
                                className="form-check-input position-static"
                                type="checkbox"
                                aria-label="Select item"
                            />
                        </div>
                        <div className="font-weight-normal flex-1 d-flex align-items-center">
                            {/* TODO!(sqs) <span className="mr-2">{threadSettings.createPullRequests ? '50%' : '0%'} complete</span>*/}
                            {filteredItemsOrError !== LOADING && !isErrorLike(filteredItemsOrError) && (
                                <ListHeaderQueryLinksNav
                                    query={query}
                                    links={[
                                        {
                                            label: 'pending',
                                            queryField: 'is',
                                            queryValues: ['pending'],
                                            count: filteredItemsOrError.items.filter(
                                                ({ status }) => status === 'pending'
                                            ).length,
                                            icon: DotsHorizontalCircleIcon,
                                        },
                                        {
                                            label: 'open',
                                            queryField: 'is',
                                            queryValues: ['open'],
                                            count: filteredItemsOrError.items.filter(({ status }) => status === 'open')
                                                .length,
                                            icon: SourcePullIcon,
                                        },
                                        {
                                            label: 'merged',
                                            queryField: 'is',
                                            queryValues: ['merged'],
                                            count: filteredItemsOrError.items.filter(
                                                ({ status }) => status === 'merged'
                                            ).length,
                                            icon: CheckCircleIcon,
                                        },
                                        {
                                            label: 'closed',
                                            queryField: 'is',
                                            queryValues: ['closed'],
                                            count: filteredItemsOrError.items.filter(
                                                ({ status }) => status === 'closed'
                                            ).length,
                                            icon: CloseCircleIcon,
                                        },
                                    ]}
                                    location={location}
                                    className="flex-1"
                                />
                            )}
                        </div>
                        {/* TODO!(sqs) <div className="d-flex">
                            <ThreadActionsPullRequestListHeaderFilterButtonDropdown
                                header="Filter by who's assigned"
                                items={['sqs (you)', 'ekonev', 'jleiner', 'ziyang', 'kting7', 'ffranksena']}
                            >
                                Assignee
                            </ThreadActionsPullRequestListHeaderFilterButtonDropdown>
                            <ThreadActionsPullRequestListHeaderFilterButtonDropdown
                                header="Sort by"
                                items={['Priority', 'Most recently updated', 'Least recently updated']}
                            >
                                Sort
                            </ThreadActionsPullRequestListHeaderFilterButtonDropdown>
                            {action}
                                </div>*/}
                        {action}
                    </div>
                    {/*{threadSettings.createPullRequests && <ThreadStatusItemsProgressBar />}*/}
                    {filteredItemsOrError === LOADING ? (
                        <LoadingSpinner className="m-2" />
                    ) : filteredItemsOrError.filteredItems.length === 0 ? (
                        <p className="p-2 mb-0 text-muted">No pull requests found.</p>
                    ) : (
                        <div className="list-group list-group-flush">
                            {filteredItemsOrError.filteredItems.map((pull, i) => (
                                <ThreadActionsPullRequestsListItem
                                    key={i}
                                    thread={thread}
                                    onThreadUpdate={onThreadUpdate}
                                    threadSettings={threadSettings}
                                    pull={pull}
                                    className="list-group-item p-2"
                                    extensionsController={extensionsController}
                                />
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    )
}
