import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ExternalServiceCard } from '../../components/ExternalServiceCard'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchRepository } from './backend'
import { ErrorAlert } from '../../components/alerts'
import { defaultExternalServices } from '../../site-admin/externalServices'

interface Props extends RouteComponentProps<{}> {
    repo: GQL.IRepository
    onDidUpdateRepository: (update: Partial<GQL.IRepository>) => void
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
 * The repository settings options page.
 */
export class RepoSettingsOptionsPage extends React.PureComponent<Props, State> {
    private repoUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        that.state = {
            loading: false,
            repo: props.repo,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettings')

        that.subscriptions.add(
            that.repoUpdates.pipe(switchMap(() => fetchRepository(that.props.repo.name))).subscribe(
                repo => that.setState({ repo }),
                err => that.setState({ error: err.message })
            )
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const services = that.state.repo.externalServices.nodes
        return (
            <div className="repo-settings-options-page">
                <PageTitle title="Repository settings" />
                <h2>Settings</h2>
                {that.state.loading && <LoadingSpinner className="icon-inline" />}
                {that.state.error && <ErrorAlert error={that.state.error} />}
                {services.length > 0 && (
                    <div className="mb-4">
                        {services.map(service => (
                            <div className="mb-3" key={service.id}>
                                <ExternalServiceCard
                                    {...defaultExternalServices[service.kind]}
                                    kind={service.kind}
                                    title={service.displayName}
                                    shortDescription="Update this external service configuration to manage repository mirroring."
                                    to={`/site-admin/external-services/${service.id}`}
                                />
                            </div>
                        ))}
                        {services.length > 1 && (
                            <p>
                                This repository is mirrored by multiple external services. To change access, disable, or
                                remove that repository, the configuration must be updated on all external services.
                            </p>
                        )}
                    </div>
                )}
                <Form>
                    <div className="form-group">
                        <label htmlFor="repo-settings-options-page__name">Repository name</label>
                        <input
                            id="repo-settings-options-page__name"
                            type="text"
                            className="form-control"
                            readOnly={true}
                            disabled={true}
                            value={that.state.repo.name}
                            required={true}
                            spellCheck={false}
                            autoCapitalize="off"
                            autoCorrect="off"
                            aria-describedby="repo-settings-options-page__name-help"
                        />
                    </div>
                </Form>
            </div>
        )
    }
}
