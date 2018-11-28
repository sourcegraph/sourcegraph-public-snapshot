import * as H from 'history'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, mergeMap } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { mutateGraphQL } from '../backend/graphql'
import { Form } from '../components/Form'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'

interface Props {
    history: H.History
    isLightTheme: boolean
}

interface State {
    input: GQL.IAddExternalServiceInput

    /**
     * Holds any error returned by the remote GraphQL endpoint on failed requests.
     */
    error?: Error

    /**
     * True if the form is currently being submitted
     */
    loading: boolean
}

const defaultConfig = `{
    // Configure your external service here (Ctrl+Space to see hints)
}`

export class SiteAdminAddExternalServicePage extends React.Component<Props, State> {
    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            loading: false,
            input: {
                kind: GQL.ExternalServiceKind.GITHUB,
                displayName: '',
                config: defaultConfig,
            },
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('AddExternalService')
        this.subscriptions.add(
            this.submits
                .pipe(
                    mergeMap(() =>
                        this.addExternalService().pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ error, loading: false })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    service => {
                        this.setState({ loading: false })
                        this.props.history.push(`/site-admin/external-services/${service.id}`)
                    },
                    error => {
                        console.error(error)
                    }
                )
        )
    }

    private addExternalService(): Observable<GQL.IExternalService> {
        return mutateGraphQL(
            gql`
                mutation addExternalService($input: AddExternalServiceInput!) {
                    addExternalService(input: $input) {
                        id
                    }
                }
            `,
            { input: this.state.input }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.addExternalService || (errors && errors.length > 0)) {
                    eventLogger.log('AddExternalServiceFailed')
                    throw createAggregateError(errors)
                }
                eventLogger.log('AddExternalServiceSucceeded', {
                    externalService: {
                        kind: data.addExternalService.kind,
                    },
                })
                return data.addExternalService
            })
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="add-external-service-page">
                <PageTitle title="Add external service" />
                <h1>Add a new external service</h1>
                <p>Sourcegraph can synchronize data (e.g. code) from external services.</p>
                <SiteAdminExternalServiceForm
                    error={this.state.error}
                    input={this.state.input}
                    history={this.props.history}
                    isLightTheme={this.props.isLightTheme}
                    mode="create"
                    loading={this.state.loading}
                    onSubmit={this.onSubmit}
                    onChange={this.onChange}
                />
            </div>
        )
    }

    private onChange = (input: GQL.IAddExternalServiceInput) => {
        this.setState({ input })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        this.setState({ loading: true })
        this.submits.next()
    }
}
