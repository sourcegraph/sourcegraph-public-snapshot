import * as GQL from '../../../shared/src/graphql/schema'
import * as LSIF from '../../../shared/src/lsif/description'
import * as React from 'react'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import { eventLogger } from '../tracking/eventLogger'
import { fetchLsifJobs, fetchLsifJobStatistics } from './backend'
import { Link } from '../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import { upperFirst } from 'lodash'
import {
    FilteredConnection,
    FilteredConnectionQueryArgs,
    FilteredConnectionFilter,
} from '../components/FilteredConnection'

interface LsifJobNodeProps {
    node: GQL.ILsifJob
}

interface LsifJobNodeState {}

class LsifJobNode extends React.PureComponent<LsifJobNodeProps, LsifJobNodeState> {
    public state: LsifJobNodeState = {}

    public render(): JSX.Element | null {
        return (
            <li className="repository-node list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <Link to={`/site-admin/lsif-jobs/${this.props.node.id}`}>
                            <strong>{LSIF.lsifJobDescription(this.props.node)}</strong>
                        </Link>
                    </div>
                    <div className="repository-node__actions">
                        {this.props.node.status === 'queued' && <TimerSandIcon className="icon-inline" />}
                        {this.props.node.status === 'scheduled' && <TimerSandIcon className="icon-inline" />}
                        {this.props.node.status === 'active' && <LoadingSpinner className="icon-inline" />}
                        {this.props.node.status === 'completed' && <CheckIcon className="icon-inline" />}
                        {this.props.node.status === 'failed' && <AlertCircleIcon className="icon-inline" />}
                    </div>
                </div>
            </li>
        )
    }
}

interface Props extends RouteComponentProps<any> {}

interface State {
    stats?: GQL.ILsifJobStats
    error?: Error
}

class FilteredLsifJobsConnection extends FilteredConnection<{}, LsifJobNodeProps> {}

/**
 * A page displaying LSIF job activity.
 */
export class SiteAdminLsifJobsPage extends React.Component<Props, State> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'Active',
            id: 'active',
            tooltip: 'Show active jobs only',
            args: { status: 'active' },
        },
        {
            label: 'Queued',
            id: 'queued',
            tooltip: 'Show queued jobs only',
            args: { status: 'queued' },
        },
        {
            label: 'Scheduled',
            id: 'scheduled',
            tooltip: 'Show scheduled jobs only',
            args: { status: 'scheduled' },
        },
        {
            label: 'Completed',
            id: 'completed',
            tooltip: 'Show completed jobs only',
            args: { status: 'completed' },
        },
        {
            label: 'Failed',
            id: 'failed',
            tooltip: 'Show failed jobs only',
            args: { status: 'failed' },
        },
    ]

    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminLsifJobPage')

        this.subscriptions.add(
            fetchLsifJobStatistics().subscribe(stats => this.setState({ stats }), error => this.setState({ error }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-lsif-queues-page">
                <PageTitle title="LSIF Jobs - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">LSIF Jobs</h2>
                </div>

                {this.state.error && <p className="alert alert-danger">{upperFirst(this.state.error.message)}</p>}
                {this.state.stats && (
                    <>
                        <div>
                            {this.state.stats.active} active, {this.state.stats.queued} queued,{' '}
                            {this.state.stats.scheduled} scheduled, {this.state.stats.completed} completed, and{' '}
                            {this.state.stats.failed} failed
                        </div>

                        <FilteredLsifJobsConnection
                            className="list-group list-group-flush mt-3"
                            noun="job"
                            pluralNoun="jobs"
                            queryConnection={this.queryJobs}
                            nodeComponent={LsifJobNode}
                            filters={SiteAdminLsifJobsPage.FILTERS}
                            history={this.props.history}
                            location={this.props.location}
                            appendResults={true}
                        />
                    </>
                )}
            </div>
        )
    }

    // TODO - do on-refresh of stats

    private queryJobs = (args: FilteredConnectionQueryArgs) => fetchLsifJobs({ ...args })
}
