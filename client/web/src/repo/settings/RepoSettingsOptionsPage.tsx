import * as React from 'react'

import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError } from '@sourcegraph/common'
import { Container, PageHeader, LoadingSpinner, Input, Text } from '@sourcegraph/wildcard'

import { ExternalServiceCard } from '../../components/externalServices/ExternalServiceCard'
import { defaultExternalServices } from '../../components/externalServices/externalServices'
import { PageTitle } from '../../components/PageTitle'
import { SettingsAreaRepositoryFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { fetchSettingsAreaRepository } from './backend'

interface Props extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
}

interface State {
    /**
     * The repository object, refreshed after we make changes that modify it.
     */
    repo: SettingsAreaRepositoryFields

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

        this.state = {
            loading: false,
            repo: props.repo,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettings')

        this.subscriptions.add(
            this.repoUpdates.pipe(switchMap(() => fetchSettingsAreaRepository(this.props.repo.name))).subscribe(
                repo => this.setState({ repo }),
                error => this.setState({ error: asError(error).message })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const services = this.state.repo.externalServices.nodes
        return (
            <>
                <PageTitle title="Repository settings" />
                <PageHeader path={[{ text: 'Settings' }]} headingElement="h2" className="mb-3" />
                <Container className="repo-settings-options-page">
                    {this.state.loading && <LoadingSpinner />}
                    {this.state.error && <ErrorAlert error={this.state.error} />}
                    {services.length > 0 && (
                        <div className="mb-3">
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
                                <Text>
                                    This repository is mirrored by multiple external services. To change access,
                                    disable, or remove this repository, the configuration must be updated on all
                                    external services.
                                </Text>
                            )}
                        </div>
                    )}
                    <Form>
                        <Input
                            id="repo-settings-options-page__name"
                            readOnly={true}
                            disabled={true}
                            value={this.state.repo.name}
                            required={true}
                            spellCheck={false}
                            autoCapitalize="off"
                            autoCorrect="off"
                            label="Repository name"
                            className="mb-0"
                        />
                    </Form>
                </Container>
            </>
        )
    }
}
