import H from 'history'
import { isEqual } from 'lodash'
import React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap, withLatestFrom } from 'rxjs/operators'
import { gql, mutateGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { Form } from '../components/Form'
import { eventLogger } from '../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { withAuthenticatedUser } from './withAuthenticatedUser'

interface Props {
    authenticatedUser: GQL.IUser
    location: H.Location
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The result of creating the token, null before it has started, loading, or an error. */
    creationOrError: GQL.ICreateAccessTokenResult | null | typeof LOADING | ErrorLike

    /** Whether the user canceled the authorization request. */
    canceled: boolean
}

/**
 * Prompts the user to approve or deny an authorization request from a client application. If approved, it
 * generates an access token that is accessible to the requesting client application.
 */
export const ClientAuthorizationFlow = withAuthenticatedUser(
    class ClientAuthorizationFlow extends React.PureComponent<Props, State> {
        public state: State = { creationOrError: null, canceled: false }

        private submits = new Subject<void>()
        private componentUpdates = new Subject<Props>()
        private subscriptions = new Subscription()

        public componentDidMount(): void {
            eventLogger.logViewEvent('ClientAuthorizationFlow')

            const argsOrError = this.componentUpdates.pipe(
                map(({ location }) => this.getCreateAccessTokenArgsFromLocation(location)),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )

            this.subscriptions.add(
                this.submits
                    .pipe(
                        withLatestFrom(argsOrError),
                        switchMap(
                            ([, argsOrError]) =>
                                isErrorLike(argsOrError)
                                    ? [asError(argsOrError)]
                                    : createAccessToken({ ...argsOrError, user: this.props.authenticatedUser.id }).pipe(
                                          catchError(error => [asError(error)]),
                                          startWith(LOADING)
                                      )
                        ),
                        map(result => ({ creationOrError: result }))
                    )
                    .subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
            )

            this.componentUpdates.next(this.props)
        }

        public componentWillReceiveProps(nextProps: Props): void {
            this.componentUpdates.next(nextProps)
        }

        public componentWillUnmount(): void {
            this.subscriptions.unsubscribe()
        }

        /** Parses the args for creating the access token from the URL. */
        private getCreateAccessTokenArgsFromLocation(
            location: H.Location
        ): Pick<GQL.ICreateAccessTokenOnMutationArguments, 'scopes' | 'note' | 'requestSession'> | ErrorLike {
            const params = new URLSearchParams(location.search)
            const note = params.get('note')
            const requestSession = params.get('request-session')
            if (note && requestSession) {
                // Only support "user:all" scope for now.
                return { scopes: ['user:all'], note, requestSession }
            }
            if (!note && !requestSession) {
                return new Error('missing required note and request-session URL query parameters')
            }
            if (!note) {
                return new Error('missing required note URL query parameter')
            }
            return new Error('missing required request-session URL query parameter')
        }

        public render(): JSX.Element | null {
            const args = this.getCreateAccessTokenArgsFromLocation(this.props.location)
            return (
                <div className="client-authorization-flow">
                    {!isErrorLike(args) ? (
                        !this.state.canceled ? (
                            <Form onSubmit={this.onSubmit}>
                                <p>
                                    Request session: <code>{args.requestSession}</code>
                                </p>
                                <p>
                                    Note: <code>{args.note}</code>
                                </p>
                                <p>
                                    Scopes: <code>{args.scopes.join(' ')}</code>
                                </p>
                                <div className="d-flex">
                                    <button type="button" className="btn btn-secondary" onClick={this.onCancel}>
                                        Cancel
                                    </button>
                                    <button type="submit" className="btn btn-success btn-lg">
                                        Allow
                                    </button>
                                </div>
                            </Form>
                        ) : (
                            <p>Canceled</p>
                        )
                    ) : (
                        <div className="alert alert-danger">Invalid client authorization request: {args.message}.</div>
                    )}
                </div>
            )
        }

        private onSubmit: React.FormEventHandler<HTMLFormElement> = e => {
            e.preventDefault()
            this.submits.next()
        }

        private onCancel: React.MouseEventHandler<HTMLButtonElement> = e => {
            e.preventDefault()
            this.setState({ canceled: true })
        }
    }
)

function createAccessToken(args: GQL.ICreateAccessTokenOnMutationArguments): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation CreateAccessTokenForClientAuthorizationFlow(
                $user: ID!
                $scopes: [String!]!
                $note: String!
                $requestSession: String!
            ) {
                createAccessToken(user: $user, scopes: $scopes, note: $note, requestSession: $requestSession) {
                    id
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.createAccessToken || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}
