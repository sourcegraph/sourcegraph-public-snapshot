import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllThreads } from './backend'

class ThreadNode extends React.PureComponent<{ node: GQL.IThread }> {
    public render(): JSX.Element | null {
        const commentsTitle = this.props.node.comments
            .map(c => `${c.author.username} on ${format(c.createdAt, 'YYYY-MM-DD')}:\n\n${c.contents}\n`)
            .join('\n-------------------------------------\n')

        const url = `/threads/${this.props.node.id}`

        return (
            <li>
                <div>
                    <Link to={url}>{this.props.node.title}</Link>
                    <br />
                    <span className="text-muted">{this.props.node.author.username}</span>
                </div>
                <ul>
                    <li title={commentsTitle}>
                        <Link to={url}>
                            {this.props.node.comments.length} {pluralize('comment', this.props.node.comments.length)}
                        </Link>
                    </li>
                    <li>Repository: {this.props.node.repo.canonicalRemoteID}</li>
                    <li>File: {this.props.node.repoRevisionPath}</li>
                    <li>
                        Branch: {this.props.node.branch} ({this.props.node.repoRevision.slice(0, 6)})
                    </li>
                    <li>Org: {this.props.node.repo.org.name}</li>
                    {this.props.node.createdAt && <li>Created: {format(this.props.node.createdAt, 'YYYY-MM-DD')}</li>}
                    {this.props.node.archivedAt && (
                        <li>Archived: {format(this.props.node.archivedAt, 'YYYY-MM-DD')}</li>
                    )}
                </ul>
            </li>
        )
    }
}

interface Props extends RouteComponentProps<any> {}

export interface State {
    threads?: GQL.IThread[]
    totalCount?: number
}

class FilteredThreadConnection extends FilteredConnection<GQL.IThread> {}

/**
 * A page displaying the threads on this site.
 */
export class SiteAdminThreadsPage extends React.Component<Props, State> {
    public state: State = {}

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminThreads')
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-threads-page">
                <PageTitle title="Threads - Admin" />
                <h2>Threads and comments (beta)</h2>
                <FilteredThreadConnection
                    className="mt-3"
                    noun="thread"
                    pluralNoun="threads"
                    queryConnection={fetchAllThreads}
                    hideFilter={true}
                    nodeComponent={ThreadNode}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }
}
