import H from 'history'
import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { useQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ListHeaderQueryLinksButtonGroup } from '../../../threadsOLD/components/ListHeaderQueryLinks'
import { ThreadsListFilter } from '../../../threadsOLD/list/ThreadsListFilter'
import { ThreadsIcon } from '../../icons'
import { ThreadListFilterContext } from '../../list/header/ThreadListFilterDropdownButton'
import { ThreadList, ThreadListHeaderCommonFilters } from '../../list/ThreadList'
import { useThreads } from '../../list/useThreads'

interface Props extends ExtensionsControllerNotificationProps {
    location: H.Location
    history: H.History
}

const QUERY_FIELDS_IN_USE = ['involves', 'author', 'mentions']

/**
 * A list of all threads.
 */
export const GlobalThreadsListPage: React.FunctionComponent<Props> = props => {
    const [query, onQueryChange] = useQueryParameter(props)
    const threads = useThreads({ query })
    const filterProps: ThreadListFilterContext = {
        threadConnection: threads,
        query,
        onQueryChange,
    }
    return (
        <>
            <h3 className="d-flex align-items-center">
                <ThreadsIcon className="icon-inline mr-1" /> All threads
            </h3>
            <div className="d-flex justify-content-between align-items-start">
                <div className="flex-1 mr-5">
                    <ListHeaderQueryLinksButtonGroup
                        query={query}
                        links={[
                            {
                                label: 'Involved',
                                queryField: 'involves',
                                queryValues: ['sqs'], // TODO!(sqs): un-hardcode
                                removeQueryFields: QUERY_FIELDS_IN_USE,
                            },
                            {
                                label: 'Created',
                                queryField: 'author',
                                queryValues: ['sqs'], // TODO!(sqs): un-hardcode
                                removeQueryFields: QUERY_FIELDS_IN_USE,
                            },
                            {
                                label: 'Mentioned',
                                queryField: 'mentions',
                                queryValues: ['sqs'], // TODO!(sqs): un-hardcode
                                removeQueryFields: QUERY_FIELDS_IN_USE,
                            },
                        ]}
                        location={props.location}
                        itemClassName="font-weight-bold px-3"
                        itemActiveClassName="btn-primary"
                        itemInactiveClassName="btn-link"
                    />
                </div>
                <div className="flex-1 mb-3">
                    <ThreadsListFilter value={query} onChange={onQueryChange} />
                </div>
            </div>
            <ThreadList
                {...props}
                threads={threads}
                query={query}
                onQueryChange={onQueryChange}
                itemCheckboxes={true}
                showRepository={true}
                headerItems={{
                    right: (
                        <>
                            <ThreadListHeaderCommonFilters {...filterProps} />
                        </>
                    ),
                }}
            />
        </>
    )
}
