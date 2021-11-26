import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { ComponentChangesFields, GitCommitFields } from '../../../../graphql-operations'
import { GitCommitNodeByline } from '../../../../repo/commits/GitCommitNodeByline'

interface Props {
    component: ComponentChangesFields
    className?: string
}

export const ComponentCommits: React.FunctionComponent<Props> = ({ component: { commits }, className }) => (
    <div className={className}>
        {commits && commits.nodes.length > 0 ? (
            <ol className={classNames('list-group list-group-flush')}>
                {commits.nodes.map(commit => (
                    <GitCommit key={commit.oid} commit={commit} tag="li" className="list-group-item py-2" />
                ))}
            </ol>
        ) : (
            <p>No changes found</p>
        )}
    </div>
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
                    <Link to={commit.canonicalURL} className="text-body" title={commit.message}>
                        {commit.subject}
                    </Link>
                </h4>
            }
            className="d-flex align-items-center small text-muted"
        />
    </Tag>
)
