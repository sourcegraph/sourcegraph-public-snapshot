import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { gql, mutateGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { Timestamp } from '../../components/time/Timestamp'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'

export const externalAccountFragment = gql`
    fragment ExternalAccountFields on ExternalAccount {
        id
        serviceType
        serviceID
        accountID
        createdAt
        updatedAt
        refreshURL
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

    onDidUpdate: () => void
}

interface ExternalAccountNodeState {
    /** Undefined means in progress, null means done or not started. */
    deletionOrError?: null | ErrorLike
}

export class ExternalAccountNode extends React.PureComponent<ExternalAccountNodeProps, ExternalAccountNodeState> {
    public state: ExternalAccountNodeState = { deletionOrError: null }

    private deletes = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.deletes
                .pipe(
                    filter(() => window.confirm('Really delete the association with this external account?')),
                    switchMap(() =>
                        deleteExternalAccount(this.props.node.id).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ deletionOrError: c })),
                            tap(() => {
                                if (this.props.onDidUpdate) {
                                    this.props.onDidUpdate()
                                }
                            }),
                            startWith<Pick<ExternalAccountNodeState, 'deletionOrError'>>({ deletionOrError: null })
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
        const loading = this.state.deletionOrError === undefined
        return (
            <li className="list-group-item py-2">
                <div className="d-flex align-items-center justify-content-between">
                    <div className="mr-2 text-truncate">
                        <span className="badge badge-secondary">{this.props.node.serviceType}</span>{' '}
                        <strong>{this.props.node.serviceID}</strong>
                        <br />
                        <span className="text-muted">
                            {this.props.node.accountID} &mdash;{' '}
                            <small>
                                updated <Timestamp date={this.props.node.updatedAt} />
                            </small>
                        </span>
                    </div>
                    <div className="text-nowrap">
                        {this.props.node.refreshURL && (
                            <a className="btn btn-secondary" href={this.props.node.refreshURL}>
                                Refresh
                            </a>
                        )}{' '}
                        <button className="btn btn-danger" onClick={this.deleteExternalAccount} disabled={loading}>
                            Delete
                        </button>
                        {isErrorLike(this.state.deletionOrError) && (
                            <div className="alert alert-danger mt-2">
                                Error: {upperFirst(this.state.deletionOrError.message)}
                            </div>
                        )}
                    </div>
                </div>
            </li>
        )
    }

    private deleteExternalAccount = () => this.deletes.next()
}
