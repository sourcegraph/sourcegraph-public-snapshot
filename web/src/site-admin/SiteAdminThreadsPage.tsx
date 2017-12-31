import format from 'date-fns/format'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { pluralize } from '../util/strings'
import { fetchAllThreads } from './backend'

interface ThreadListItemProps {
    className: string

    /**
     * The thread to display in this list item.
     */
    thread: GQL.IThread

    /**
     * Called when the thread is updated by an action in this list item.
     */
    onDidUpdate?: () => void
}

interface ThreadListItemState {
    loading: boolean
    errorDescription?: string
}

class ThreadListItem extends React.PureComponent<ThreadListItemProps, ThreadListItemState> {
    public state: ThreadListItemState = {
        loading: false,
    }

    public render(): JSX.Element | null {
        const commentsTitle = this.props.thread.comments
            .map(c => `${c.author.username} on ${format(c.createdAt, 'YYYY-MM-DD')}:\n\n${c.contents}\n`)
            .join('\n-------------------------------------\n')

        return (
            <li className={this.props.className}>
                <div className="site-admin-detail-list__header">
                    {this.props.thread.title}
                    <br />
                    <span className="site-admin-detail-list__display-name">{this.props.thread.author.username}</span>
                </div>
                <ul className="site-admin-detail-list__info">
                    <li title={commentsTitle}>
                        {this.props.thread.comments.length} {pluralize('comment', this.props.thread.comments.length)}
                    </li>
                    <li>ID: {this.props.thread.id}</li>
                    <li>Repository: {this.props.thread.repo.canonicalRemoteID}</li>
                    <li>File: {this.props.thread.repoRevisionPath}</li>
                    <li>
                        Branch: {this.props.thread.branch} ({this.props.thread.repoRevision.slice(0, 6)})
                    </li>
                    <li>Org: {this.props.thread.repo.org.name}</li>
                    {this.props.thread.createdAt && (
                        <li>Created: {format(this.props.thread.createdAt, 'YYYY-MM-DD')}</li>
                    )}
                    {this.props.thread.archivedAt && (
                        <li>Archived: {format(this.props.thread.archivedAt, 'YYYY-MM-DD')}</li>
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

        this.subscriptions.add(
            this.threadUpdates
                .pipe(mergeMap(fetchAllThreads))
                .subscribe(resp => this.setState({ threads: resp.nodes, totalCount: resp.totalCount }))
        )
        this.threadUpdates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list site-admin-threads-page">
                <PageTitle title="Threads - Admin" />
                <h2>
                    Threads{' '}
                    {typeof this.state.totalCount === 'number' &&
                        this.state.totalCount > 0 &&
                        `(${this.state.totalCount})`}
                </h2>
                <ul className="site-admin-detail-list__list">
                    {this.state.threads &&
                        this.state.threads.map(thread => (
                            <ThreadListItem
                                key={thread.id}
                                className="site-admin-detail-list__item"
                                thread={thread}
                                onDidUpdate={this.onDidUpdateThread}
                            />
                        ))}
                </ul>
                {this.state.threads && this.state.threads.length === 0 && <p>No threads to display.</p>}
                {this.state.threads &&
                    typeof this.state.totalCount === 'number' &&
                    this.state.totalCount > 0 && (
                        <p>
                            <small>
                                {this.state.totalCount} {pluralize('thread', this.state.totalCount)} total{' '}
                                {this.state.threads.length < this.state.totalCount &&
                                    `(showing first ${this.state.threads.length})`}
                            </small>
                        </p>
                    )}
            </div>
        )
    }

    private onDidUpdateThread = () => this.threadUpdates.next()
}
