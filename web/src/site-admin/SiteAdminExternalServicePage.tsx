import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, mapTo, mergeMap, startWith, switchMap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { queryGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ExternalServiceCard } from './../components/ExternalServiceCard'
import { getExternalService } from './externalServices'
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
    updatedOrError: null | true | typeof LOADING | ErrorLike
}

export class SiteAdminExternalServicePage extends React.Component<Props, State> {
    public state: State = {
        externalServiceOrError: LOADING,
        updatedOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private submits = new Subject<GQL.IUpdateExternalServiceInput>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminExternalService')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(props => props.match.params.id),
                    distinctUntilChanged(),
                    switchMap(id =>
                        fetchExternalService(id).pipe(
                            startWith(LOADING),
                            catchError(err => [asError(err)])
                        )
                    ),
                    map(result => ({ externalServiceOrError: result }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(input =>
                        concat(
                            [{ updatedOrError: LOADING }],
                            updateExternalService(input).pipe(
                                mapTo(null),
                                mergeMap(() =>
                                    concat(
                                        // Flash "updated" text
                                        of({ updatedOrError: true }),
                                        // Hide "updated" text again after 1s
                                        of({ updatedOrError: null }).pipe(delay(1000))
                                    )
                                ),
                                catchError((error: Error) => [{ updatedOrError: asError(error) }])
                            )
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public render(): JSX.Element | null {
        let error: ErrorLike | undefined
        if (isErrorLike(this.state.updatedOrError)) {
            error = this.state.updatedOrError
        }

        const externalService =
            (!isErrorLike(this.state.externalServiceOrError) &&
                this.state.externalServiceOrError !== LOADING &&
                this.state.externalServiceOrError) ||
            undefined

        const externalServiceCategory = externalService && getExternalService(externalService.kind)

        return (
            <div className="site-admin-configuration-page mt-3">
                {externalService ? (
                    <PageTitle title={`External service - ${externalService.displayName}`} />
                ) : (
                    <PageTitle title="External service" />
                )}
                <h2>Update external service</h2>
                {this.state.externalServiceOrError === LOADING && <LoadingSpinner className="icon-inline" />}
                {isErrorLike(this.state.externalServiceOrError) && (
                    <p className="alert alert-danger">{upperFirst(this.state.externalServiceOrError.message)}</p>
                )}
                {externalService && (
                    <div className="mb-3">
                        <ExternalServiceCard {...getExternalService(externalService.kind)} />
                    </div>
                )}
                {externalService && externalServiceCategory && (
                    <SiteAdminExternalServiceForm
                        input={externalService}
                        editorActions={externalServiceCategory.editorActions}
                        jsonSchema={externalServiceCategory.jsonSchema}
                        error={error}
                        mode="edit"
                        loading={this.state.updatedOrError === LOADING}
                        onSubmit={this.onSubmit}
                        onChange={this.onChange}
                        history={this.props.history}
                        isLightTheme={this.props.isLightTheme}
                    />
                )}
                {this.state.updatedOrError === true && (
                    <p className="alert alert-success user-settings-profile-page__alert">Updated!</p>
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
