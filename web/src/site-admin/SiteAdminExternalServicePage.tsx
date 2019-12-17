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
import { CodeHostCard } from '../components/CodeHostCard'
import { getCodeHost } from './externalServices'
import { SiteAdminCodeHostForm } from './SiteAdminCodeHostForm'
import { ErrorAlert } from '../components/alerts'

interface Props extends RouteComponentProps<{ id: GQL.ID }> {
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

interface State {
    externalServiceOrError: typeof LOADING | GQL.ICodeHost | ErrorLike

    /**
     * The result of updating the external service: null when complete or not started yet,
     * loading, or an error.
     */
    updatedOrError: null | true | typeof LOADING | ErrorLike

    warning?: string
}

export class SiteAdminCodeHostPage extends React.Component<Props, State> {
    public state: State = {
        externalServiceOrError: LOADING,
        updatedOrError: null,
    }

    private componentUpdates = new Subject<Props>()
    private submits = new Subject<GQL.IUpdateCodeHostInput>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminCodeHost')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(props => props.match.params.id),
                    distinctUntilChanged(),
                    switchMap(id =>
                        fetchCodeHost(id).pipe(
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
                            [{ updatedOrError: LOADING, warning: null }],
                            updateCodeHost(input).pipe(
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

        const externalServiceCategory = externalService && getCodeHost(externalService.kind)

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
                    <ErrorAlert className="mb-3" error={this.state.externalServiceOrError} />
                )}
                {externalService && (
                    <div className="mb-3">
                        <CodeHostCard
                            {...getCodeHost(externalService.kind)}
                            kind={externalService.kind}
                        />
                    </div>
                )}
                {externalService && externalServiceCategory && (
                    <SiteAdminCodeHostForm
                        input={externalService}
                        editorActions={externalServiceCategory.editorActions}
                        jsonSchema={externalServiceCategory.jsonSchema}
                        error={error}
                        warning={this.state.warning}
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

    private onChange = (input: GQL.IAddCodeHostInput): void => {
        this.setState(state => {
            if (isCodeHost(state.externalServiceOrError)) {
                return { ...state, externalServiceOrError: { ...state.externalServiceOrError, ...input } }
            }
            return state
        })
    }

    private onSubmit = (event?: React.FormEvent<HTMLFormElement>): void => {
        if (event) {
            event.preventDefault()
        }
        if (isCodeHost(this.state.externalServiceOrError)) {
            this.submits.next(this.state.externalServiceOrError)
        }
    }
}

function isCodeHost(
    externalServiceOrError: typeof LOADING | GQL.ICodeHost | ErrorLike
): externalServiceOrError is GQL.ICodeHost {
    return externalServiceOrError !== LOADING && !isErrorLike(externalServiceOrError)
}

function updateCodeHost(
    input: GQL.IUpdateCodeHostInput
): Observable<Pick<GQL.ICodeHost, 'warning'>> {
    return mutateGraphQL(
        gql`
            mutation UpdateCodeHost($input: UpdateCodeHostInput!) {
                updateCodeHost(input: $input) {
                    warning
                }
            }
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateCodeHost)
    )
}

function fetchCodeHost(id: GQL.ID): Observable<GQL.ICodeHost> {
    return queryGraphQL(
        gql`
            query CodeHost($id: ID!) {
                node(id: $id) {
                    ... on CodeHost {
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
        map(data => data.node as GQL.ICodeHost)
    )
}
