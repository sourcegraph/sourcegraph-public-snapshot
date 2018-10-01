import * as H from 'history'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Observable } from 'rxjs'
import { replaceRevisionInURL } from '.'
import * as GQL from '../backend/graphqlschema'
import { fetchCommits } from '../commits/backend'
import { FilteredConnection } from '../components/FilteredConnection'
import { UserAvatar } from '../user/UserAvatar'

interface CommitNodeProps {
    node: GQL.IGitCommit
    location: H.Location
}

const CommitNode: React.SFC<CommitNodeProps> = ({ node, location }) => (
    <li className="repo-rev-sidebar-commits-node">
        <NavLink
            to={replaceRevisionInURL(location.pathname + location.search + location.hash, node.oid as string)}
            className="repo-rev-sidebar-commits-node__link"
            activeClassName="repo-rev-sidebar-commits-node__link--active"
        >
            {node.author.person ? (
                <UserAvatar user={node.author.person} tooltip={node.author.person.name} className="icon-inline mr-1" />
            ) : (
                <UserAvatar className="icon-inline mr-1" />
            )}
            <span className="repo-rev-sidebar-commits-node__name">{node.abbreviatedOID}</span>
            <span className="repo-rev-sidebar-commits-node__message">
                <small>{node.message}</small>
            </span>
        </NavLink>
    </li>
)

interface Props {
    repoID: GQL.ID
    rev: string | undefined
    filePath: string
    history: H.History
    location: H.Location
}

export class RepoRevSidebarCommits extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <FilteredConnection<GQL.IGitCommit, Pick<CommitNodeProps, 'location'>>
                className="repo-rev-sidebar-commits"
                compact={true}
                noun="commit"
                pluralNoun="commits"
                queryConnection={this.fetchCommits}
                nodeComponent={CommitNode}
                nodeComponentProps={{ location: this.props.location } as Pick<CommitNodeProps, 'location'>}
                defaultFirst={100}
                shouldUpdateURLQuery={false}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private fetchCommits = (args: { query?: string }): Observable<GQL.IGitCommitConnection> =>
        fetchCommits(this.props.repoID, this.props.rev || '', { ...args, currentPath: this.props.filePath || '' })
}
