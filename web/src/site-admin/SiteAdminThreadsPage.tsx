import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { FilteredConnection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllThreads } from './backend'

interface ThreadNodeProps {
    /**
     * The thread to display in this list item.
     */
    node: GQL.IThread

    /**
     * Called when the thread is updated by an action in this list item.
     */
    onDidUpdate?: () => void
}

interface ThreadNodeState {
    loading: boolean
    errorDescription?: string
}

class ThreadNode extends React.PureComponent<ThreadNodeProps, ThreadNodeState> {
    public state: ThreadNodeState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        const commentsTitle = this.props.node.comments
            .map(c => `${c.author.username} on ${format(c.createdAt, 'YYYY-MM-DD')}:\n\n${c.contents}\n`)
            .join('\n-------------------------------------\n')

        const url = `/threads/${this.props.node.id}`

        return (
            <li className="site-admin-detail-list__item">
                <div className="site-admin-detail-list__header">
                    <Link to={url}>{this.props.node.title}</Link>
                    <br />
                    <span className="site-admin-detail-list__display-name">{this.props.node.author.username}</span>
                </div>
                <ul className="site-admin-detail-list__info">
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
                <div className="site-admin-detail-list__actions">
                    {this.state.errorDescription && (
                        <p className="site-admin-detail-list__error">{this.state.errorDescription}</p>
                    )}
                </div>
            </li>
        )
    }
}

interface Props extends RouteComponentProps<any> {}

export interface State {
    threads?: GQL.IThread[]
    totalCount?: number
}

/**
 * A page displaying the threads on this site.
 */
export class SiteAdminThreadsPage extends React.Component<Props, State> {
    public state: State = {}

    private threadUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminThreads')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<ThreadNodeProps, 'onDidUpdate'> = {
            onDidUpdate: this.onDidUpdateThread,
        }

        return (
            <div className="site-admin-detail-list site-admin-threads-page">
                <PageTitle title="Threads - Admin" />
                <h2>Threads and comments (beta)</h2>
                <p>
                    Code comments are in beta and require{' '}
                    <a href="https://about.sourcegraph.com/products/editor">Sourcegraph Editor</a>.
                </p>
                <FilteredConnection
                    className="site-admin-page__filtered-connection"
                    noun="thread"
                    pluralNoun="threads"
                    queryConnection={fetchAllThreads}
                    hideFilter={true}
                    nodeComponent={ThreadNode}
                    nodeComponentProps={nodeProps}
                    updates={this.threadUpdates}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private onDidUpdateThread = () => this.threadUpdates.next()
}
