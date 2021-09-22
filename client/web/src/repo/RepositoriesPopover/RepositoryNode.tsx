import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { RepositoryPopoverFields } from '../../graphql-operations'

interface RepositoryNodeProps {
    node: RepositoryPopoverFields
    currentRepo?: Scalars['ID']
}

export const RepositoryNode: React.FunctionComponent<RepositoryNodeProps> = ({ node, currentRepo }) => (
    <li key={node.id} className="connection-popover__node">
        <Link
            to={`/${node.name}`}
            className={classNames(
                'connection-popover__node-link',
                node.id === currentRepo && 'connection-popover__node-link--active'
            )}
        >
            {displayRepoName(node.name)}
        </Link>
    </li>
)
