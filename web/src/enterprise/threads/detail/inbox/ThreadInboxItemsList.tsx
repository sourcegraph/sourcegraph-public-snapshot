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
import { QueryParameterProps } from '../../components/withQueryParameter/WithQueryParameter'
import { ThreadSettings } from '../../settings'
import { DiagnosticInfo, getDiagnosticInfos } from '../backend'
import { ThreadInboxFileItem } from './item/ThreadInboxFileItem'
import { diagnosticFilter, ThreadInboxSidebar } from './sidebar/ThreadInboxSidebar'

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
 * The list of thread inbox items.
 */
export const ThreadInboxItemsList: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    query,
    onQueryChange,
    className = '',
    extensionsController,
    ...props
}) => {
    const [diagnosticsOrError, setDiagnosticsOrError] = useState<typeof LOADING | DiagnosticInfo[] | ErrorLike>(LOADING)
    // tslint:disable-next-line: no-floating-promises
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            getDiagnosticInfos(extensionsController)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setDiagnosticsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [thread.id, extensionsController])

    return (
        <div className={`thread-inbox-items-list ${className}`}>
            {isErrorLike(diagnosticsOrError) ? (
                <div className="alert alert-danger mt-2">{diagnosticsOrError.message}</div>
            ) : (
                <>
                    {diagnosticsOrError !== LOADING &&
                        !isErrorLike(diagnosticsOrError) &&
                        /* TODO!(sqs) <WithStickyTop scrollContainerSelector=".thread-area">
                            {({ isStuck }) => (
                                <ThreadInboxItemsNavbar
                                    {...props}
                                    thread={thread}
                                    onThreadUpdate={onThreadUpdate}
                                    threadSettings={threadSettings}
                                    items={diagnosticsOrError}
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
                    {diagnosticsOrError === LOADING ? (
                        <LoadingSpinner className="mt-2" />
                    ) : diagnosticsOrError.length === 0 ? (
                        <p className="p-4 mb-0 text-muted">Inbox is empty.</p>
                    ) : (
                        <div className="d-flex">
                            <Resizable
                                className="sticky-top border-right"
                                handlePosition="right"
                                storageKey="thread-inbox-items-list__sidebar-resizable"
                                defaultSize={216 /* px */}
                                element={
                                    <ThreadInboxSidebar
                                        diagnostics={diagnosticsOrError}
                                        query={query}
                                        onQueryChange={onQueryChange}
                                        className="flex-1"
                                    />
                                }
                                style={{
                                    minWidth: '8rem',
                                    maxWidth: '75vh',
                                    height: 'calc(100vh - 83.5px)', // 83.5px = 39px + 44.5px(GlobalNavbar)
                                    top: '39px', // TODO!(sqs): this is the hardcoded height of ThreadAreaNavbar
                                }}
                            />
                            <ul className="list-unstyled mb-0 flex-1" style={{ minWidth: '0' }}>
                                {diagnosticsOrError
                                    .filter(diagnostic => diagnosticFilter(query, diagnostic))
                                    .map((diagnostic, i) => (
                                        <li key={i}>
                                            <ThreadInboxFileItem
                                                {...props}
                                                key={i}
                                                thread={thread}
                                                threadSettings={threadSettings}
                                                diagnostic={diagnostic}
                                                onThreadUpdate={onThreadUpdate}
                                                className="m-2"
                                                headerClassName="thread-inbox-items-list__item-header sticky-top"
                                                headerStyle={{
                                                    // TODO!(sqs): this is the hardcoded height of ThreadAreaNavbar
                                                    top: '39px',
                                                }}
                                                extensionsController={extensionsController}
                                            />
                                        </li>
                                    ))}
                            </ul>
                        </div>
                    )}
                </>
            )}
        </div>
    )
}
