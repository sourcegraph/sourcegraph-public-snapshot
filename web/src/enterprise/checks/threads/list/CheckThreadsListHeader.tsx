import H from 'history'
import React from 'react'
import { ListHeaderQueryLinksButtonGroup } from '../../../threads/components/ListHeaderQueryLinks'
import { QueryParameterProps } from '../../../threads/components/withQueryParameter/WithQueryParameter'
import { ThreadsAreaContext } from '../../../threads/global/ThreadsArea'
import { ThreadsListFilter } from '../../../threads/list/ThreadsListFilter'

interface Props extends QueryParameterProps, Pick<ThreadsAreaContext, 'type' | 'authenticatedUser'> {
    location: H.Location
}

const QUERY_FIELDS_IN_USE = ['involves', 'author', 'mentions']

/**
 * The header for the list of check threads.
 */
export const CheckThreadsListHeader: React.FunctionComponent<Props> = ({
    authenticatedUser,
    query,
    onQueryChange,
    location,
}) => (
    <div className="d-flex justify-content-between mb-3">
        {authenticatedUser && (
            <div className="mr-5 pr-5">
                <ListHeaderQueryLinksButtonGroup
                    query={query}
                    links={[
                        {
                            label: 'Involved',
                            queryField: 'involves',
                            queryValues: [authenticatedUser.username],
                            removeQueryFields: QUERY_FIELDS_IN_USE,
                        },
                        {
                            label: 'Created',
                            queryField: 'author',
                            queryValues: [authenticatedUser.username],
                            removeQueryFields: QUERY_FIELDS_IN_USE,
                        },
                    ]}
                    location={location}
                    itemClassName="font-weight-bold px-3"
                    itemActiveClassName="btn-primary"
                    itemInactiveClassName="btn-link"
                />
            </div>
        )}
        <div className="flex-1">
            <ThreadsListFilter value={query} onChange={onQueryChange} />
        </div>
    </div>
)
