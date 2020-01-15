import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../backend/graphql'
import { Timestamp } from '../../../components/time/Timestamp'
import { userURL } from '../../../user'
import { ErrorAlert } from '../../../components/alerts'

export const externalAccountFragment = gql`
    fragment ExternalAccountFields on ExternalAccount {
        id
        user {
            id
            username
        }
        serviceType
        serviceID
        clientID
        accountID
        createdAt
        updatedAt
        refreshURL
        accountData
    }
`

function deleteExternalAccount(externalAccount: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteExternalAccount($externalAccount: ID!) {
                deleteExternalAccount(externalAccount: $externalAccount) {
                    alwaysNil
                }
            }
        `,
        { externalAccount }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.deleteExternalAccount || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}

export interface ExternalAccountNodeProps {
    node: GQL.IExternalAccount

    showUser: boolean

    onDidUpdate: () => void
}

interface ExternalAccountNodeState {
    /** Undefined means in progress, null means done or not started. */
    deletionOrError?: null | ErrorLike

    showData: boolean
}

export class ExternalAccountNode extends React.PureComponent<ExternalAccountNodeProps, ExternalAccountNodeState> {
    public state: ExternalAccountNodeState = {
        deletionOrError: null,
        showData: false,
    }

    private deletes = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        that.subscriptions.add(
            that.deletes
                .pipe(
                    filter(() => window.confirm('Really delete the association with this external account?')),
                    switchMap(() =>
                        deleteExternalAccount(that.props.node.id).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ deletionOrError: c })),
                            tap(() => {
                                if (that.props.onDidUpdate) {
                                    that.props.onDidUpdate()
                                }
                            }),
                            startWith<Pick<ExternalAccountNodeState, 'deletionOrError'>>({ deletionOrError: undefined })
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
        const loading = that.state.deletionOrError === undefined
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div className="mr-2 text-truncate">
                        {that.props.showUser && (
                            <>
                                <strong>
                                    <Link to={userURL(that.props.node.user.username)}>
                                        {that.props.node.user.username}
                                    </Link>
                                </strong>{' '}
                                &mdash;{' '}
                            </>
                        )}
                        <span className="badge badge-secondary">{that.props.node.serviceType}</span>{' '}
                        {that.props.node.accountID}
                        {(that.props.node.serviceID || that.props.node.clientID) && (
                            <small className="text-muted">
                                <br />
                                {that.props.node.serviceID}
                                {that.state.showData && that.props.node.clientID && (
                                    <> &mdash; {that.props.node.clientID}</>
                                )}
                            </small>
                        )}
                        <br />
                        <small className="text-muted">
                            Updated <Timestamp date={that.props.node.updatedAt} />
                        </small>
                    </div>
                    <div className="text-nowrap">
                        {that.props.node.accountData && (
                            <button type="button" className="btn btn-secondary" onClick={that.toggleShowData}>
                                {that.state.showData ? 'Hide' : 'Show'} data
                            </button>
                        )}{' '}
                        {that.props.node.refreshURL && (
                            <a className="btn btn-secondary" href={that.props.node.refreshURL}>
                                Refresh
                            </a>
                        )}{' '}
                        <button
                            type="button"
                            className="btn btn-danger"
                            onClick={that.deleteExternalAccount}
                            disabled={loading}
                        >
                            Delete
                        </button>
                        {isErrorLike(that.state.deletionOrError) && (
                            <ErrorAlert className="mt-2" error={that.state.deletionOrError} />
                        )}
                    </div>
                </div>
                {that.state.showData && (
                    <pre className="p-2 mt-2 mb-4">
                        <small>{JSON.stringify(that.props.node.accountData, null, 2)}</small>
                    </pre>
                )}
            </li>
        )
    }

    private deleteExternalAccount = (): void => that.deletes.next()

    private toggleShowData = (): void => that.setState(prev => ({ showData: !prev.showData }))
}
