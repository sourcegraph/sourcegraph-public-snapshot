import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertOutlineIcon from 'mdi-react/AlertOutlineIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { ListHeaderQueryLinksNav } from '../components/ListHeaderQueryLinks'
import { WithThreadsQueryResults } from '../components/withThreadsQueryResults/WithThreadsQueryResults'
import { ThreadsAreaContext } from '../global/ThreadsArea'
import { threadNoun } from '../util'
import { ThreadsListItem } from './ThreadsListItem'

export interface ThreadsListContext {
    /**
     * Whether each item should have a checkbox.
     */
    itemCheckboxes?: boolean
}

interface Props extends QueryParameterProps, ThreadsListContext, Pick<ThreadsAreaContext, 'type'> {
    /**
     * A React fragment to render on the right side of the list header.
     */
    rightHeaderFragment?: React.ReactFragment

    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The list of threads with a header.
 */
export const ThreadsList: React.FunctionComponent<Props> = ({
    type,
    rightHeaderFragment,
    itemCheckboxes,
    query,
    onQueryChange,
    ...props
}) => (
    <WithThreadsQueryResults query={query}>
        {({ threadsOrError }) => (
            <div className="threads-list">
                {isErrorLike(threadsOrError) ? (
                    <div className="alert alert-danger mt-2">{threadsOrError.message}</div>
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
                            {threadsOrError !== LOADING ? (
                                <ListHeaderQueryLinksNav
                                    query={query}
                                    links={[
                                        {
                                            label: 'open',
                                            queryField: 'is',
                                            queryValues: [type.toLowerCase(), 'open'],
                                            count: threadsOrError.totalCount,
                                            icon: AlertOutlineIcon,
                                        },
                                        {
                                            label: 'closed',
                                            queryField: 'is',
                                            queryValues: [type.toLowerCase(), 'closed'],
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
                        {threadsOrError === LOADING ? (
                            <LoadingSpinner className="m-3" />
                        ) : threadsOrError.nodes.length === 0 ? (
                            <p className="p-2 mb-0 text-muted">No {threadNoun(type, true)} found.</p>
                        ) : (
                            <ul className="list-group list-group-flush">
                                {threadsOrError.nodes.map((thread, i) => (
                                    <ThreadsListItem
                                        key={i}
                                        location={props.location}
                                        thread={thread}
                                        itemCheckboxes={itemCheckboxes}
                                    />
                                ))}
                            </ul>
                        )}
                    </div>
                )}
            </div>
        )}
    </WithThreadsQueryResults>
)
