import React from 'react'
import { Link } from 'react-router-dom'

import { CatalogComponentChangesFields, GitCommitFields } from '../../../../../graphql-operations'
import { GitCommitNodeByline } from '../../../../../repo/commits/GitCommitNodeByline'

interface Props {
    catalogComponent: CatalogComponentChangesFields
    className?: string
    headerClassName?: string
    titleClassName?: string
}

export const ComponentChanges: React.FunctionComponent<Props> = ({
    catalogComponent: { editCommits },
    className,
    headerClassName,
    titleClassName,
}) =>
    editCommits && editCommits.nodes.length > 0 ? (
        <div className={className}>
            <header className={headerClassName}>
                <h3 className={titleClassName}>Changes</h3>
            </header>
            <ol className="list-group list-group-flush">
                {editCommits.nodes.map(commit => (
                    <GitCommit key={commit.oid} commit={commit} tag="li" className="list-group-item py-2" />
                ))}
            </ol>
        </div>
    ) : (
        <p>No changes found</p>
    )

const GitCommit: React.FunctionComponent<{ commit: GitCommitFields; tag: 'li'; className?: string }> = ({
    commit,
    tag: Tag,
    className,
}) => (
    <Tag className={className}>
        <GitCommitNodeByline
            author={commit.author}
            committer={commit.committer}
            messageElement={
                <h4 className="h6 mb-0 text-truncate">
                    <Link to={commit.canonicalURL} className="text-body">
                        {commit.message}
                    </Link>
                </h4>
            }
            className="d-flex align-items-center small text-muted"
        />
    </Tag>
)
