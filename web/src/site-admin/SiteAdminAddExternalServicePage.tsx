import * as H from 'history'
import React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { mutateGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { refreshSiteFlags } from '../site/backend'
import { ThemeProps } from '../theme'
import { ExternalServiceCard } from './ExternalServiceCard'
import { ExternalServiceVariant, getExternalService } from './externalServices'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'

interface Props extends ThemeProps {
    history: H.History
    kind: ExternalServiceKind
    variant?: ExternalServiceVariant
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

/**
 * Page for adding a single external service
 */
export class SiteAdminAddExternalServicePage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        const serviceKindMetadata = getExternalService(this.props.kind, this.props.variant)
        this.state = {
            loading: false,
            displayName: serviceKindMetadata.defaultDisplayName,
            config: serviceKindMetadata.defaultConfig,
        }
    }

    private submits = new Subject<GQL.IAddExternalServiceInput>()
    private subscriptions = new Subscription()

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
        const externalService = getExternalService(this.props.kind, this.props.variant)
        return (
            <div className="add-external-service-page mt-3">
                <PageTitle title="Add external service" />
                <h1>Add external service</h1>
                <div className="mb-3">
                    <ExternalServiceCard {...externalService} />
                </div>
                <div className="mb-4">{externalService.longDescription}</div>
                <SiteAdminExternalServiceForm
                    {...this.props}
                    error={this.state.error}
                    input={this.getExternalServiceInput()}
                    editorActions={externalService.editorActions}
                    jsonSchema={externalService.jsonSchema}
                    mode="create"
                    onSubmit={this.onSubmit}
                    onChange={this.onChange}
                    loading={this.state.loading}
                />
            </div>
        )
    }

    private getExternalServiceInput(): GQL.IAddExternalServiceInput {
        return {
            displayName: this.state.displayName,
            config: this.state.config,
            kind: this.props.kind,
        }
    }

    private onChange = (input: GQL.IAddExternalServiceInput) => {
        this.setState({
            displayName: input.displayName,
            config: input.config,
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
