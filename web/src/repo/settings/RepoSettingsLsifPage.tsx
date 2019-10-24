import * as GQL from '../../../../shared/src/graphql/schema'
import * as React from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Subscription, Subject } from 'rxjs'
import { upperFirst } from 'lodash'
import { fetchLsifDumps } from './backend'
import { switchMap, tap } from 'rxjs/operators'

const DumpRef: React.FunctionComponent<{
    dumpRef: GQL.ILSIFDump
}> = ({ dumpRef }) => {
    return (
        <li className="repo-settings-lsif-page__ref">
            {dumpRef.commit}@{dumpRef.root}
        </li>
    )
}

interface Props extends RouteComponentProps<any> {
    repo: GQL.IRepository
}

interface State {
    dumps?: GQL.ILSIFDumpConnection
    loading: boolean
    error?: Error
}

/**
 * The repository settings LSIF page.
 */
export class RepoSettingsLsifPage extends React.PureComponent<Props, State> {
    public state: State = { loading: true }

    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsLsif')

        this.subscriptions.add(
            this.updates
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(() => fetchLsifDumps(this.props.repo.name))
                )
                .subscribe(
                    dumps => this.setState({ dumps, loading: false }),
                    error => this.setState({ error, loading: false })
                )
        )
        this.updates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repo-settings-lsif-page">
                <PageTitle title="LSIF" />
                <h2>LSIF</h2>
                {this.state.loading && <LoadingSpinner className="icon-inline" />}
                {this.state.error && <div className="alert alert-danger">{upperFirst(this.state.error.message)}</div>}
                {this.state.dumps && (
                    <ul className="repo-settings-index-page__refs">
                        {this.state.dumps.nodes.map((ref, i) => (
                            <DumpRef key={i} dumpRef={ref} />
                        ))}
                    </ul>
                )}
            </div>
        )
    }
}
