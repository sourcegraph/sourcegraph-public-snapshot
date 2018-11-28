import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { queryGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'

interface Props extends RouteComponentProps<{ id: GQL.ID }> {
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

interface State {
    externalServiceOrError: typeof LOADING | GQL.IExternalService | ErrorLike

    /**
     * The result of updating the external service: null when complete or not started yet,
     * loading, or an error.
     */
    updateOrError: null | typeof LOADING | ErrorLike
}

export class SiteAdminExternalServicePage extends React.Component<Props, State> {
    public state: State = {
        externalServiceOrError: LOADING,
        updateOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private submits = new Subject<GQL.IUpdateExternalServiceInput>()
    private subscriptions = new Subscription()

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminExternalService')

        const externalServiceChanges = this.componentUpdates.pipe(
            map(props => props.match.params.id),
            distinctUntilChanged(),
            switchMap(id =>
                fetchExternalService(id).pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
            )
        )

        this.subscriptions.add(
            externalServiceChanges
                .pipe(map(result => ({ externalServiceOrError: result })))
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(input =>
                        updateExternalService(input).pipe(
                            mapTo(null),
                            startWith(LOADING)
                        )
                    ),
                    catchError(err => [asError(err)]),
                    map(u => ({ updateOrError: u }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let error: ErrorLike | undefined
        if (isErrorLike(this.state.updateOrError)) {
            error = this.state.updateOrError
        }

        let externalService: GQL.IExternalService | undefined
        if (isErrorLike(this.state.externalServiceOrError)) {
            error = this.state.externalServiceOrError
        } else if (this.state.externalServiceOrError !== LOADING) {
            externalService = this.state.externalServiceOrError
        }

        const loading = this.state.updateOrError === LOADING || this.state.externalServiceOrError === LOADING
        return (
            <div className="site-admin-configuration-page">
                <PageTitle title="External Service - " />
                <h2>External Service</h2>
                {externalService && (
                    <SiteAdminExternalServiceForm
                        input={externalService}
                        error={error}
                        mode="edit"
                        loading={loading}
                        onSubmit={this.onSubmit}
                        onChange={this.onChange}
                        history={this.props.history}
                        isLightTheme={this.props.isLightTheme}
                    />
                )}
            </div>
        )
    }

    private onChange = (input: GQL.IAddExternalServiceInput) => {
        this.setState(state => {
            if (isExternalService(state.externalServiceOrError)) {
                return { ...state, externalServiceOrError: { ...state.externalServiceOrError, ...input } }
            }
            return state
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>) => {
        if (event) {
            event.preventDefault()
        }
        if (isExternalService(this.state.externalServiceOrError)) {
            this.submits.next(this.state.externalServiceOrError)
        }
    }
}

function isExternalService(
    externalServiceOrError: typeof LOADING | GQL.IExternalService | ErrorLike
): externalServiceOrError is GQL.IExternalService {
    return externalServiceOrError !== LOADING && !isErrorLike(externalServiceOrError)
}

function updateExternalService(input: GQL.IUpdateExternalServiceInput): Observable<GQL.IExternalService> {
    return queryGraphQL(
        gql`
            mutation UpdateExternalService($input: UpdateExternalServiceInput!) {
                updateExternalService(input: $input) {
                    id
                    kind
                    displayName
                    config
                }
            }
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node as GQL.IExternalService)
    )
}

function fetchExternalService(id: GQL.ID): Observable<GQL.IExternalService> {
    return queryGraphQL(
        gql`
            query ExternalService($id: ID!) {
                node(id: $id) {
                    ... on ExternalService {
                        id
                        kind
                        displayName
                        config
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node as GQL.IExternalService)
    )
}
