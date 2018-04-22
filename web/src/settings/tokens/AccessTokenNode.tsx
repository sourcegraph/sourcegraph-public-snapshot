import { upperFirst } from 'lodash'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { catchError } from 'rxjs/operators/catchError'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mapTo } from 'rxjs/operators/mapTo'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, mutateGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { FilteredConnection } from '../../components/FilteredConnection'
import { Timestamp } from '../../components/time/Timestamp'
import { userURL } from '../../user'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { AccessTokenCreatedAlert } from './AccessTokenCreatedAlert'

export const accessTokenFragment = gql`
    fragment AccessTokenFields on AccessToken {
        id
        note
        createdAt
        lastUsedAt
        user {
            username
        }
    }
`

function deleteAccessToken(tokenID: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteAccessToken($tokenID: ID!) {
                deleteAccessToken(byID: $tokenID) {
                    alwaysNil
                }
            }
        `,
        { tokenID }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.deleteAccessToken || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}

export interface AccessTokenNodeProps {
    node: GQL.IAccessToken

    /** Whether the user who owns the token should be displayed. */
    showUser?: boolean

    /**
     * The newly created token, if any. This contains the secret for this node's token iff node.id
     * === newToken.id.
     */
    newToken?: GQL.ICreateAccessTokenResult

    onDidUpdate: () => void
}

export interface AccessTokenNodeState {
    /** Undefined means in progress, null means done or not started. */
    deletionOrError?: null | ErrorLike
}

export class AccessTokenNode extends React.PureComponent<AccessTokenNodeProps, AccessTokenNodeState> {
    public state: AccessTokenNodeState = { deletionOrError: null }

    private deletes = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.deletes
                .pipe(
                    filter(() =>
                        window.confirm(
                            'Really delete and revoke this token? Any clients using it will no longer be able to access the Sourcegraph API.'
                        )
                    ),
                    switchMap(() =>
                        deleteAccessToken(this.props.node.id).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ deletionOrError: c })),
                            tap(() => {
                                if (this.props.onDidUpdate) {
                                    this.props.onDidUpdate()
                                }
                            }),
                            startWith<Pick<AccessTokenNodeState, 'deletionOrError'>>({ deletionOrError: null })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const note = this.props.node.note || '(no description)'
        const loading = this.state.deletionOrError === undefined
        return (
            <li className="list-group-item p-3 d-block">
                <div className="d-flex w-100 justify-content-between">
                    <div className="mr-2">
                        {this.props.showUser ? (
                            <>
                                <strong>
                                    <Link to={userURL(this.props.node.user.username)}>
                                        {this.props.node.user.username}
                                    </Link>
                                </strong>{' '}
                                &mdash; {note}
                            </>
                        ) : (
                            <strong>{note}</strong>
                        )}{' '}
                        <small className="text-muted">
                            {' '}
                            &mdash;{' '}
                            {this.props.node.lastUsedAt ? (
                                <>
                                    last used <Timestamp date={this.props.node.lastUsedAt} />
                                </>
                            ) : (
                                'never used'
                            )}, created <Timestamp date={this.props.node.createdAt} />
                        </small>
                    </div>
                    <div>
                        <button className="btn btn-danger" onClick={this.deleteAccessToken} disabled={loading}>
                            Delete
                        </button>
                        {isErrorLike(this.state.deletionOrError) && (
                            <div className="alert alert-danger mt-2">
                                Error: {upperFirst(this.state.deletionOrError.message)}
                            </div>
                        )}
                    </div>
                </div>
                {this.props.newToken &&
                    this.props.node.id === this.props.newToken.id && (
                        <AccessTokenCreatedAlert
                            className="alert alert-success mt-4"
                            token={this.props.newToken.token}
                        />
                    )}
            </li>
        )
    }

    private deleteAccessToken = () => this.deletes.next()
}

export class FilteredAccessTokenConnection extends FilteredConnection<
    GQL.IAccessToken,
    Pick<AccessTokenNodeProps, 'onDidUpdate'>
> {}
