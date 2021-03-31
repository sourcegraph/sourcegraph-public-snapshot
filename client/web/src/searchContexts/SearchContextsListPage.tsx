import React, { useCallback } from 'react'
import * as H from 'history'
import { Link } from '../../../shared/src/components/Link'
import { FilteredConnection } from '../components/FilteredConnection'
import { Page } from '../components/Page'
import { ListSearchContextsResult, ListSearchContextsVariables, SearchContextFields } from '../graphql-operations'
import { fetchAutoDefinedSearchContexts, fetchSearchContexts } from '../search/backend'
import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

function getSearchContextRepositoriesDescription(searchContext: ISearchContext): string {
    const numberRepos = searchContext.repositories.length
    return searchContext.autoDefined ? 'Auto-defined' : `${numberRepos} repositor${numberRepos === 1 ? 'y' : 'ies'}`
}

export interface SearchContextNodeProps {
    node: SearchContextFields
    location: H.Location
    history: H.History
}

export const SearchContextNode: React.FunctionComponent<SearchContextNodeProps> = ({
    node,
    location,
}: SearchContextNodeProps) => (
    <div className="search-context-node card mb-1 p-3">
        {node.autoDefined ? <div>{node.spec}</div> : <Link to={`${location.pathname}/${node.id}`}>{node.spec}</Link>}
        <div>
            {getSearchContextRepositoriesDescription(node as ISearchContext)} &middot; {node.description}
        </div>
    </div>
)

export interface SearchContextsListPageProps {
    location: H.Location
    history: H.History
}

export const SearchContextsListPage: React.FunctionComponent<SearchContextsListPageProps> = props => {
    const queryConnection = useCallback(
        (args: Partial<ListSearchContextsVariables>) =>
            fetchSearchContexts(args.first ?? 1, args.query ?? undefined, args.after ?? undefined),
        []
    )

    const autoDefinedSearchContexts = useObservable(fetchAutoDefinedSearchContexts)

    return (
        <div className="w-100">
            <Page>
                <div className="search-contexts-list-page">
                    <div className="search-contexts-list-page__title mb-3">
                        <h2>Search contexts</h2>
                    </div>

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
                </div>
            </Page>
        </div>
    )
}
