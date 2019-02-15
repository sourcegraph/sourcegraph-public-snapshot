import * as H from 'history'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { mutateGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'
import { ALL_EXTERNAL_SERVICES, ExternalServiceMetadata } from './externalServices'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'

interface Props {
    history: H.History
    location: H.Location
    isLightTheme: boolean
    eventLogger: {
        logViewEvent: (event: 'AddExternalService') => void
        log: (event: 'AddExternalServiceFailed' | 'AddExternalServiceSucceeded', eventProperties?: any) => void
    }
}

interface State {
    displayName: string
    config: string

    /**
     * Holds any error returned by the remote GraphQL endpoint on failed requests.
     */
    error?: Error

    /**
     * True if the form is currently being submitted
     */
    loading: boolean
}

export class SiteAdminAddExternalServicePage extends React.Component<Props, State> {
    public state: State = {
        loading: false,
        displayName: '',
        config: this.getExternalServiceMetadata().defaultConfig,
    }

    private submits = new Subject<GQL.IAddExternalServiceInput>()
    private subscriptions = new Subscription()

    private getExternalServiceKind(kind?: string | GQL.ExternalServiceKind): GQL.ExternalServiceKind {
        if (!kind) {
            const params = new URLSearchParams(this.props.history.location.search)
            kind = params.get('kind') || undefined
        }
        if (kind) {
            kind = kind.toUpperCase()
        }
        const isKnownKind = (kind: string): kind is GQL.ExternalServiceKind =>
            !!ALL_EXTERNAL_SERVICES[kind as GQL.ExternalServiceKind]
        return kind && isKnownKind(kind) ? kind : GQL.ExternalServiceKind.GITHUB // default to GitHub
    }

    private getExternalServiceMetadata(kind?: string | GQL.ExternalServiceKind): ExternalServiceMetadata {
        return ALL_EXTERNAL_SERVICES[this.getExternalServiceKind(kind)]
    }

    private getExternalServiceInput(): GQL.IAddExternalServiceInput {
        return {
            displayName: this.state.displayName,
            config: this.state.config,
            kind: this.getExternalServiceKind(),
        }
    }

    public componentDidMount(): void {
        this.props.eventLogger.logViewEvent('AddExternalService')
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(input =>
                        addExternalService(input, this.props.eventLogger).pipe(
                            map(() => {
                                // Refresh site flags so that global site alerts
                                // reflect the latest configuration.
                                refreshSiteFlags().subscribe(undefined, err => console.error(err))

                                this.setState({ loading: false })
                                this.props.history.push(`/site-admin/external-services`)
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
        if (input.kind.toLowerCase() === this.getExternalServiceKind().toLowerCase()) {
            this.setState({
                displayName: input.displayName,
                config: input.config,
            })
            return
        }

        this.setState({
            displayName: input.displayName,
            config: this.getExternalServiceMetadata(input.kind).defaultConfig,
        })

        const { search, ...loc } = this.props.location

        const params = new URLSearchParams(search)
        params.set('kind', input.kind.toLowerCase())

        this.props.history.replace({
            ...loc,
            search: params.toString(),
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        this.submits.next(this.getExternalServiceInput())
    }
}

function addExternalService(
    input: GQL.IAddExternalServiceInput,
    eventLogger: Pick<Props['eventLogger'], 'log'>
): Observable<GQL.IExternalService> {
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
