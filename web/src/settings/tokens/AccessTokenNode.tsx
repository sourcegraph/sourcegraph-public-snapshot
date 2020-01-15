import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { Timestamp } from '../../components/time/Timestamp'
import { userURL } from '../../user'
import { AccessTokenCreatedAlert } from './AccessTokenCreatedAlert'
import { ErrorAlert } from '../../components/alerts'

export const accessTokenFragment = gql`
    fragment AccessTokenFields on AccessToken {
        id
        scopes
        note
        createdAt
        lastUsedAt
        subject {
            username
        }
        creator {
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

    /** Whether the token's subject user should be displayed. */
    showSubject?: boolean

    /**
     * The newly created token, if any. This contains the secret for this node's token iff node.id
     * === newToken.id.
     */
    newToken?: GQL.ICreateAccessTokenResult

    onDidUpdate: () => void
}

interface AccessTokenNodeState {
    /** Undefined means in progress, null means done or not started. */
    deletionOrError?: null | ErrorLike
}

export class AccessTokenNode extends React.PureComponent<AccessTokenNodeProps, AccessTokenNodeState> {
    public state: AccessTokenNodeState = { deletionOrError: null }

    private deletes = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        that.subscriptions.add(
            that.deletes
                .pipe(
                    filter(() =>
                        window.confirm(
                            'Delete and revoke this token? Any clients using it will no longer be able to access the Sourcegraph API.'
                        )
                    ),
                    switchMap(() =>
                        deleteAccessToken(that.props.node.id).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ deletionOrError: c })),
                            tap(() => {
                                if (that.props.onDidUpdate) {
                                    that.props.onDidUpdate()
                                }
                            }),
                            startWith<Pick<AccessTokenNodeState, 'deletionOrError'>>({ deletionOrError: undefined })
                        )
                    )
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    error => console.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const note = that.props.node.note || '(no description)'
        const loading = that.state.deletionOrError === undefined
        return (
            <li className="list-group-item p-3 d-block" data-e2e-access-token-description={that.props.node.note}>
                <div className="d-flex w-100 justify-content-between">
                    <div className="mr-2">
                        {that.props.showSubject ? (
                            <>
                                <strong>
                                    <Link to={userURL(that.props.node.subject.username)}>
                                        {that.props.node.subject.username}
                                    </Link>
                                </strong>{' '}
                                &mdash; {note}
                            </>
                        ) : (
                            <strong>{note}</strong>
                        )}{' '}
                        <small className="text-muted">
                            {' '}
                            &mdash; <em>{that.props.node.scopes && that.props.node.scopes.join(', ')}</em>
                            <br />
                            {that.props.node.lastUsedAt ? (
                                <>
                                    Last used <Timestamp date={that.props.node.lastUsedAt} />
                                </>
                            ) : (
                                'Never used'
                            )}
                            , created <Timestamp date={that.props.node.createdAt} />
                            {that.props.node.subject.username !== that.props.node.creator.username && (
                                <>
                                    {' '}
                                    by{' '}
                                    <Link to={userURL(that.props.node.creator.username)}>
                                        {that.props.node.creator.username}
                                    </Link>
                                </>
                            )}
                        </small>
                    </div>
                    <div>
                        <button
                            type="button"
                            className="btn btn-danger e2e-access-token-delete"
                            onClick={that.deleteAccessToken}
                            disabled={loading}
                        >
                            Delete
                        </button>
                        {isErrorLike(that.state.deletionOrError) && (
                            <ErrorAlert className="mt-2" error={that.state.deletionOrError} />
                        )}
                    </div>
                </div>
                {that.props.newToken && that.props.node.id === that.props.newToken.id && (
                    <AccessTokenCreatedAlert
                        className="alert alert-success mt-4"
                        tokenSecret={that.props.newToken.token}
                        token={that.props.node}
                    />
                )}
            </li>
        )
    }

    private deleteAccessToken = (): void => that.deletes.next()
}

export class FilteredAccessTokenConnection extends FilteredConnection<
    GQL.IAccessToken,
    Pick<AccessTokenNodeProps, 'onDidUpdate'>
> {}
