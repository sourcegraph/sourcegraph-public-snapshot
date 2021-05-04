import * as H from 'history'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { Timestamp } from '../components/time/Timestamp'
import { SearchContextFields } from '../graphql-operations'

export interface SearchContextNodeProps {
    node: SearchContextFields
    location: H.Location
    history: H.History
}

export const SearchContextNode: React.FunctionComponent<SearchContextNodeProps> = ({
    node,
}: SearchContextNodeProps) => (
    <div className="search-context-node pb-4 pt-4 d-flex align-items-center">
        <div className="search-context-node__left flex-grow-1">
            <Link to={`/contexts/${node.id}`}>{node.spec}</Link>
            {node.description.length > 0 && (
                <div className="text-muted search-context-node__left__description mt-1">{node.description}</div>
            )}
        </div>
        <div className="search-context-node__right text-muted d-flex">
            <div className="mr-2">{node.repositories.length} repositories</div>
            <div>
                Updated <Timestamp date={node.updatedAt} noAbout={true} />
            </div>
        </div>
    </div>
)
