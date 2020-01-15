import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, mergeMap, startWith, switchMap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ExternalServiceCard } from '../components/ExternalServiceCard'
import { SiteAdminExternalServiceForm } from './SiteAdminExternalServiceForm'
import { ErrorAlert } from '../components/alerts'
import { defaultExternalServices } from './externalServices'

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

    warning?: string
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

        that.subscriptions.add(
            that.componentUpdates
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
                .subscribe(stateUpdate => that.setState(stateUpdate))
        )

        that.subscriptions.add(
            that.submits
                .pipe(
                    switchMap(input =>
                        concat(
                            [{ updatedOrError: LOADING, warning: null }],
                            updateExternalService(input).pipe(
                                mergeMap(({ warning }) =>
                                    warning
                                        ? of({ warning, updatedOrError: null })
                                        : concat(
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
                .subscribe(stateUpdate => that.setState(stateUpdate as State))
        )

        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public render(): JSX.Element | null {
        let error: ErrorLike | undefined
        if (isErrorLike(that.state.updatedOrError)) {
            error = that.state.updatedOrError
        }

        const externalService =
            (!isErrorLike(that.state.externalServiceOrError) &&
                that.state.externalServiceOrError !== LOADING &&
                that.state.externalServiceOrError) ||
            undefined

        const externalServiceCategory = externalService && defaultExternalServices[externalService.kind]

        return (
            <div className="site-admin-configuration-page mt-3">
                {externalService ? (
                    <PageTitle title={`External service - ${externalService.displayName}`} />
                ) : (
                    <PageTitle title="External service" />
                )}
                <h2>Update synced repositories</h2>
                {that.state.externalServiceOrError === LOADING && <LoadingSpinner className="icon-inline" />}
                {isErrorLike(that.state.externalServiceOrError) && (
                    <ErrorAlert className="mb-3" error={that.state.externalServiceOrError} />
                )}
                {externalService && (
                    <div className="mb-3">
                        <ExternalServiceCard {...defaultExternalServices[externalService.kind]} />
                    </div>
                )}
                {externalService && externalServiceCategory && (
                    <SiteAdminExternalServiceForm
                        input={externalService}
                        editorActions={externalServiceCategory.editorActions}
                        jsonSchema={externalServiceCategory.jsonSchema}
                        error={error}
                        warning={that.state.warning}
                        mode="edit"
                        loading={that.state.updatedOrError === LOADING}
                        onSubmit={that.onSubmit}
                        onChange={that.onChange}
                        history={that.props.history}
                        isLightTheme={that.props.isLightTheme}
                    />
                )}
                {that.state.updatedOrError === true && (
                    <p className="alert alert-success user-settings-profile-page__alert">Updated!</p>
                )}
            </div>
        )
    }

    private onChange = (input: GQL.IAddExternalServiceInput): void => {
        that.setState(state => {
            if (isExternalService(state.externalServiceOrError)) {
                return { ...state, externalServiceOrError: { ...state.externalServiceOrError, ...input } }
            }
            return state
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        if (isExternalService(that.state.externalServiceOrError)) {
            that.submits.next(that.state.externalServiceOrError)
        }
    }
}

function isExternalService(
    externalServiceOrError: typeof LOADING | GQL.IExternalService | ErrorLike
): externalServiceOrError is GQL.IExternalService {
    return externalServiceOrError !== LOADING && !isErrorLike(externalServiceOrError)
}

function updateExternalService(
    input: GQL.IUpdateExternalServiceInput
): Observable<Pick<GQL.IExternalService, 'warning'>> {
    return mutateGraphQL(
        gql`
            mutation UpdateExternalService($input: UpdateExternalServiceInput!) {
                updateExternalService(input: $input) {
                    warning
                }
            }
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateExternalService)
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
