import React from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { SearchContextMinimalFields } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { Badge, Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'

import styles from './SearchContextNode.module.scss'

export interface SearchContextNodeProps {
    node: SearchContextMinimalFields
    location: H.Location
    history: H.History
}

export const SearchContextNode: React.FunctionComponent<React.PropsWithChildren<SearchContextNodeProps>> = ({
    node,
}: SearchContextNodeProps) => (
    <div className={classNames('py-3 d-flex align-items-center', styles.searchContextNode)}>
        <div className={classNames('flex-grow-1', styles.left)}>
            <div>
                <Link to={`/contexts/${node.spec}`}>
                    <strong>{node.spec}</strong>
                </Link>
                {!node.public && (
                    <Badge variant="secondary" pill={true} className="ml-1" as="div">
                        Private
                    </Badge>
                )}
            </div>
            {node.query.length > 0 && (
                <small>
                    <SyntaxHighlightedSearchQuery query={node.query} key={node.name} />
                </small>
            )}

            {node.description.length > 0 && (
                <div className={classNames('text-muted mt-1', styles.leftDescription)}>{node.description}</div>
            )}
        </div>
        <div className={classNames('text-muted d-flex', styles.right)}>
            {node.repositories && node.repositories.length > 0 && (
                <div className="mr-2">{node.repositories.length} repositories</div>
            )}
            <div>
                Updated <Timestamp date={node.updatedAt} noAbout={true} />
            </div>
        </div>
    </div>
)
