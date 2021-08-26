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
    <div className="search-context-node py-3 d-flex align-items-center">
        <div className="search-context-node__left flex-grow-1">
            <div>
                <Link to={`/contexts/${node.spec}`}>
                    <strong>{node.spec}</strong>
                </Link>
                {!node.public && <div className="badge badge-pill badge-secondary ml-1">Private</div>}
            </div>

            {node.description.length > 0 && (
                <div className="text-muted search-context-node__left-description mt-1">{node.description}</div>
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
