import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import { Resizable } from '../../../../../../shared/src/components/Resizable'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThreadSettings } from '../../settings'
import { Changeset, computeChangesets } from '../backend'
import { ThreadChangesetItem } from './item/ThreadChangesetItem'
import { ThreadInboxSidebar } from './sidebar/ThreadChangesSidebar'

interface Props extends QueryParameterProps, ExtensionsControllerProps, PlatformContextProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'title' | 'type' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    className?: string
    history: H.History
    location: H.Location
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * The list of thread changes.
 */
export const ThreadChangesList: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    query,
    onQueryChange,
    className = '',
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
    }, [thread.id, threadSettings, extensionsController, thread])

    return (
        <div className={`thread-changes-list ${className}`}>
            {isErrorLike(changesetsOrError) ? (
                <div className="alert alert-danger mt-2">{changesetsOrError.message}</div>
            ) : (
                <>
                    {changesetsOrError !== LOADING &&
                        !isErrorLike(changesetsOrError) &&
                        /* TODO!(sqs) <WithStickyTop scrollContainerSelector=".thread-area">
                            {({ isStuck }) => (
                                <ThreadInboxItemsNavbar
                                    {...props}
                                    thread={thread}
                                    onThreadUpdate={onThreadUpdate}
                                    threadSettings={threadSettings}
                                    items={changesetsOrError}
                                    query={query}
                                    onQueryChange={onQueryChange}
                                    includeThreadInfo={isStuck}
                                    className={`sticky-top position-sticky row bg-body thread-inbox-items-list__navbar py-2 px-3 ${
                                        isStuck ? 'border-bottom shadow' : ''
                                    }`}
                                    extensionsController={extensionsController}
                                />
                            )}
                                </WithStickyTop>*/ ''}
                    {changesetsOrError === LOADING ? (
                        <LoadingSpinner className="m-2" />
                    ) : changesetsOrError.length === 0 ? (
                        <p className="p-4 mb-0 text-muted">Inbox is empty.</p>
                    ) : (
                        <ul className="list-unstyled mb-0 flex-1" style={{ minWidth: '0' }}>
                            {changesetsOrError.map((changeset, i) => (
                                <li key={i}>
                                    <ThreadChangesetItem
                                        key={i}
                                        className="m-2"
                                        threadSettings={threadSettings}
                                        changeset={changeset}
                                        headerClassName="thread-changes-list__item-header sticky-top"
                                        headerStyle={{
                                            // TODO!(sqs): this is the hardcoded height of ThreadAreaNavbar
                                            top: '39px',
                                        }}
                                    />
                                </li>
                            ))}
                        </ul>
                    )}
                </>
            )}
        </div>
    )
}
