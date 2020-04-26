import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchRepository } from './backend'
import { ErrorAlert } from '../../components/alerts'
import { asError } from '../../../../shared/src/util/errors'

interface Props extends RouteComponentProps<{}> {
    repo: GQL.IRepository
}

interface State {
    /**
     * The repository object, refreshed after we make changes that modify it.
     */
    repo: GQL.IRepository

    loading: boolean
    error?: string
}

/**
 * The repository settings permissions page.
 */
export class RepoSettingsPermissionsPage extends React.PureComponent<Props, State> {
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            loading: false,
            repo: props.repo,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsPermissions')

        this.subscriptions.add(
            this.updates.pipe(switchMap(() => fetchRepository(this.props.repo.name))).subscribe(
                repo => this.setState({ repo }),
                err => this.setState({ error: asError(err).message })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repo-settings-permissions-page">
                <PageTitle title="Permissions" />
                <h2>Permissions</h2>
                {this.state.loading ? (
                    <LoadingSpinner className="icon-inline" />
                ) : this.state.error ? (
                    <ErrorAlert prefix="Error getting repository permissions" error={this.state.error} />
                ) : this.state.repo.isPrivate ? (
                    this.state.repo.permissionsInfo ? (
                        <>
                            <table className="table repo-settings-permissions-page__stats">
                                <tbody>
                                    <tr>
                                        <th>Last complete sync</th>
                                        <td>
                                            {/* TODO: Fix "Unsafe member access .syncedAt on an any value" */}
                                            {this.state.repo.permissionsInfo.syncedAt ? (
                                                <Timestamp date={this.state.repo.permissionsInfo.syncedAt} />
                                            ) : (
                                                'Never'
                                            )}
                                        </td>
                                        <td className="text-muted">Updated by repository-centric syncing</td>
                                    </tr>
                                    <tr>
                                        <th>Last incremental sync</th>
                                        <td>
                                            {/* TODO: Fix "Unsafe member access .syncedAt on an any value" */}
                                            <Timestamp date={this.state.repo.permissionsInfo.updatedAt} />
                                        </td>
                                        <td className="text-muted">Updated by user-centric syncing</td>
                                    </tr>
                                </tbody>
                            </table>
                        </>
                    ) : (
                        <div className="alert alert-info">
                            This repository is queued to sync permissions, only site admins will have access to it until
                            syncing is finished.
                        </div>
                    )
                ) : (
                    <div className="alert alert-info">This is a public repository and can be viewed by everyone.</div>
                )}
            </div>
        )
    }
}
