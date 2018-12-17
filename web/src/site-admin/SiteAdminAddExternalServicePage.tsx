import * as H from 'history'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { mutateGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ALL_EXTERNAL_SERVICES } from './externalServices'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'

interface Props {
    history: H.History
    isLightTheme: boolean
}

interface State {
    input: Pick<GQL.IAddExternalServiceInput, 'displayName' | 'config'>

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
    public state: State = {
        loading: false,
        input: {
            displayName: '',
            config: defaultConfig,
        },
    }

    private submits = new Subject<GQL.IAddExternalServiceInput>()
    private subscriptions = new Subscription()

    private getExternalServiceKind(): GQL.ExternalServiceKind {
        const params = new URLSearchParams(this.props.history.location.search)
        const kind = params.get('kind')
        if (kind) {
            const service = ALL_EXTERNAL_SERVICES.find(s => s.kind === kind.toUpperCase())
            if (service) {
                return service.kind
            }
        }
        return GQL.ExternalServiceKind.GITHUB
    }

    private getExternalServiceInput(): GQL.IAddExternalServiceInput {
        return { ...this.state.input, kind: this.getExternalServiceKind() }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('AddExternalService')
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(input =>
                        addExternalService(input).pipe(
                            map(externalService => {
                                this.setState({ loading: false })
                                this.props.history.push(`/site-admin/external-services/${externalService.id}`)
                            }),
                            catchError(error => {
                                console.error(error)
                                this.setState({ error, loading: false })
                                return []
                            })
                        )
                    )
                )
                .subscribe()
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="add-external-service-page">
                <PageTitle title="Add external service" />
                <h1>Add a new external service</h1>
                <p>Sourcegraph can synchronize data (e.g. code) from external services.</p>
                <SiteAdminExternalServiceForm
                    error={this.state.error}
                    input={this.getExternalServiceInput()}
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
        this.submits.next(this.getExternalServiceInput())
    }
}

function addExternalService(input: GQL.IAddExternalServiceInput): Observable<GQL.IExternalService> {
    return mutateGraphQL(
        gql`
            mutation addExternalService($input: AddExternalServiceInput!) {
                addExternalService(input: $input) {
                    id
                }
            }
        `,
        { input }
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
