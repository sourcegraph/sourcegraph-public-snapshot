import * as H from 'history'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'

import { SearchContextFields } from '../graphql-operations'

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
