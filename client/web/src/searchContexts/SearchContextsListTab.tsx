import * as H from 'history'
import React, { useCallback } from 'react'

import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { FilteredConnection } from '../components/FilteredConnection'
import { ListSearchContextsResult, ListSearchContextsVariables, SearchContextFields } from '../graphql-operations'
import { VersionContext } from '../schema/site.schema'
import { fetchAutoDefinedSearchContexts, fetchSearchContexts } from '../search/backend'

import { SearchContextNode, SearchContextNodeProps } from './SearchContextNode'

export interface SearchContextsListTabProps {
    location: H.Location
    history: H.History
    availableVersionContexts: VersionContext[] | undefined
}

export const SearchContextsListTab: React.FunctionComponent<SearchContextsListTabProps> = props => {
    const queryConnection = useCallback(
        (args: Partial<ListSearchContextsVariables>) =>
            fetchSearchContexts({
                first: args.first ?? 1,
                query: args.query ?? undefined,
                after: args.after ?? undefined,
            }),
        []
    )

    const autoDefinedSearchContexts = useObservable(fetchAutoDefinedSearchContexts)

    return (
        <>
            <div className="mb-3">
                <h3>Auto-defined</h3>
                {autoDefinedSearchContexts?.map(context => (
                    <SearchContextNode
                        key={context.id}
                        node={context}
                        location={props.location}
                        history={props.history}
                    />
                ))}
            </div>

            <h3>User-defined</h3>
            <FilteredConnection<
                SearchContextFields,
                Omit<SearchContextNodeProps, 'node'>,
                ListSearchContextsResult['searchContexts']
            >
                history={props.history}
                location={props.location}
                defaultFirst={10}
                queryConnection={queryConnection}
                hideSearch={false}
                nodeComponent={SearchContextNode}
                nodeComponentProps={{
                    location: props.location,
                    history: props.history,
                }}
                noun="search context"
                pluralNoun="search contexts"
                noSummaryIfAllNodesVisible={true}
                cursorPaging={true}
            />
        </>
    )
}
