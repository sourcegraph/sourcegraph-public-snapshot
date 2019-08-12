import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertOutlineIcon from 'mdi-react/AlertOutlineIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { ListHeaderQueryLinksNav } from '../../threadsOLD/components/ListHeaderQueryLinks'
import { threadNoun } from '../../threadsOLD/util'
import { ThreadListItem, ThreadListItemContext } from './ThreadListItem'

export interface ThreadListContext extends ThreadListItemContext {
    /**
     * Whether each item should have a checkbox.
     */
    itemCheckboxes?: boolean
}

const LOADING: 'loading' = 'loading'

interface Props extends QueryParameterProps, ThreadListContext {
    threads: typeof LOADING | GQL.IThreadConnection | ErrorLike

    /**
     * A React fragment to render on the right side of the list header.
     */
    rightHeaderFragment?: React.ReactFragment

    history: H.History
    location: H.Location
}

/**
 * The list of threads with a header.
 */
export const ThreadList2: React.FunctionComponent<Props> = ({
    threads,
    rightHeaderFragment,
    itemCheckboxes,
    query,
    onQueryChange,
    ...props
}) => (
    <div className="thread-list">
        {isErrorLike(threads) ? (
            <div className="alert alert-danger mt-2">{threads.message}</div>
        ) : (
            <div className="card">
                <div className="card-header d-flex align-items-center justify-content-between font-weight-normal">
                    {itemCheckboxes && (
                        <div className="form-check mx-2">
                            <input
                                className="form-check-input position-static"
                                type="checkbox"
                                aria-label="Select item"
                            />
                        </div>
                    )}
                    {threads !== LOADING ? (
                        <ListHeaderQueryLinksNav
                            query={query}
                            links={[
                                {
                                    label: 'open',
                                    queryField: 'is',
                                    queryValues: ['open'],
                                    count: threads.totalCount,
                                    icon: AlertOutlineIcon,
                                },
                                {
                                    label: 'closed',
                                    queryField: 'is',
                                    queryValues: ['closed'],
                                    count: 0,
                                    icon: CheckIcon,
                                },
                            ]}
                            location={props.location}
                            className="flex-1 nav"
                        />
                    ) : (
                        <div className="flex-1" />
                    )}
                    {rightHeaderFragment && <div>{rightHeaderFragment}</div>}
                </div>
                {threads === LOADING ? (
                    <LoadingSpinner className="m-3" />
                ) : threads.nodes.length === 0 ? (
                    <p className="p-2 mb-0 text-muted">No threads found.</p>
                ) : (
                    <ul className="list-group list-group-flush">
                        {threads.nodes.map((thread, i) => (
                            <ThreadListItem key={i} {...props} thread={thread} itemCheckboxes={itemCheckboxes} />
                        ))}
                    </ul>
                )}
            </div>
        )}
    </div>
)
