import AddIcon from '@sourcegraph/icons/lib/Add'
import CircleCheckmarkIcon from '@sourcegraph/icons/lib/CircleCheckmark'
import CopyIcon from '@sourcegraph/icons/lib/Copy'
import copy from 'copy-to-clipboard'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mapTo } from 'rxjs/operators/mapTo'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, mutateGraphQL, queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { UserAreaPageProps } from '../area/UserArea'

interface AccessTokenCreatedAlertProps {
    className: string
    token: string
}

interface AccessTokenCreatedAlertState {
    flashCopiedToClipboardMessage?: boolean
}

class AccessTokenCreatedAlert extends React.PureComponent<AccessTokenCreatedAlertProps, AccessTokenCreatedAlertState> {
    public state: AccessTokenCreatedAlertState = {}

    public render(): JSX.Element | null {
        return (
            <div className={`access-token-created-alert ${this.props.className}`}>
                <p className="mt-2">
                    <code className="mr-2">{this.props.token}</code>
                    <button
                        type="button"
                        className="btn btn-primary btn-sm"
                        onClick={this.copyToClipboard}
                        disabled={this.state.flashCopiedToClipboardMessage}
                    >
                        <CopyIcon className="icon-inline" />{' '}
                        {this.state.flashCopiedToClipboardMessage ? 'Copied to clipboard!' : 'Copy to clipboard'}
                    </button>
                </p>
                <p>
                    <CircleCheckmarkIcon className="icon-inline" /> Copy your new personal access token now. You won't
                    be able to see it again.
                </p>
                <h5 className="mt-4">
                    <strong>Example usage</strong>
                </h5>
                <pre className="mt-1">
                    <code>{curlExampleCommand(this.props.token)}</code>
                </pre>
            </div>
        )
    }

    private copyToClipboard = (): void => {
        eventLogger.log('AccessTokenCopiedToClipboard')
        copy(this.props.token)
        this.setState({ flashCopiedToClipboardMessage: true })

        setTimeout(() => {
            this.setState({ flashCopiedToClipboardMessage: false })
        }, 1500)
    }
}

function curlExampleCommand(token: string): string {
    return `curl \\
  -H 'Authorization: token ${token}' \\
  -d '{"query":"query { currentUser { username } }"}' \\
  ${window.context.appURL}/.api/graphql`
}

function deleteAccessToken(tokenID: GQLID): Observable<void> {
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

interface AccessTokenNodeProps {
    node: GQL.IAccessToken
    user: GQL.IUser

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

class AccessTokenNode extends React.PureComponent<AccessTokenNodeProps, AccessTokenNodeState> {
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
                    switchMap(() => {
                        type PartialStateUpdate = Pick<AccessTokenNodeState, 'deletionOrError'>
                        const result = deleteAccessToken(this.props.node.id).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ deletionOrError: c })),
                            tap(() => {
                                if (this.props.onDidUpdate) {
                                    this.props.onDidUpdate()
                                }
                            }),
                            publishReplay<PartialStateUpdate>(),
                            refCount()
                        )
                        return merge(of({ deletionOrError: null }), result)
                    })
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
            <li className="list-group-item p-3 d-block">
                <div className="d-flex w-100 justify-content-between">
                    <div className="mr-2">
                        <strong>{this.props.node.note || '(no description)'}</strong>{' '}
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
                        <button className="btn btn-outline-danger" onClick={this.deleteAccessToken} disabled={loading}>
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

class FilteredAccessTokenConnection extends FilteredConnection<
    GQL.IAccessToken,
    Pick<AccessTokenNodeProps, 'user' | 'onDidUpdate'>
> {}

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    /**
     * The newly created token, if any. This component must call onDidPresentNewToken
     * when it is finished presenting the token secret to the user.
     */
    newToken?: GQL.ICreateAccessTokenResult

    /**
     * Called when the newly created access token has been presented to the user and may be purged
     * from all state (and not displayed to the user anymore).
     */
    onDidPresentNewToken: () => void
}

interface State {}

/**
 * Displays a user's session token.
 */
export class UserSettingsTokensPage extends React.PureComponent<Props, State> {
    private static clearNewTokenTimer: number | undefined = undefined

    public state: State = {}

    private accessTokenUpdates = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsTokens')

        if (UserSettingsTokensPage.clearNewTokenTimer !== undefined) {
            clearTimeout(UserSettingsTokensPage.clearNewTokenTimer)
        }
    }

    public componentWillUnmount(): void {
        // Clear the newly created access token value from our application state; we assume the user
        // has already stored it elsewhere.
        this.props.onDidPresentNewToken()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<AccessTokenNodeProps, 'user' | 'onDidUpdate' | 'newToken'> = {
            user: this.props.user,
            onDidUpdate: this.onDidUpdateAccessToken,
            newToken: this.props.newToken,
        }

        return (
            <div className="user-settings-tokens-page">
                <PageTitle title="Access tokens" />
                <div className="d-flex justify-content-between align-items-center">
                    <h2>Personal access tokens</h2>
                    <Link className="btn btn-primary ml-2" to={`${this.props.match.url}/tokens/new`}>
                        <AddIcon className="icon-inline" /> Generate new token
                    </Link>
                </div>
                <p>Tokens may be used to access the Sourcegraph API.</p>
                <FilteredAccessTokenConnection
                    listClassName="list-group list-group-flush"
                    noun="access token"
                    pluralNoun="access tokens"
                    queryConnection={this.queryAccessTokens}
                    nodeComponent={AccessTokenNode}
                    nodeComponentProps={nodeProps}
                    updates={this.accessTokenUpdates}
                    hideFilter={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryAccessTokens = (args: { first?: number }): Observable<GQL.IAccessTokenConnection> =>
        queryGraphQL(
            gql`
                query AccessTokens($user: ID!, $first: Int) {
                    node(id: $user) {
                        ... on User {
                            accessTokens(first: $first) {
                                nodes {
                                    id
                                    note
                                    createdAt
                                    lastUsedAt
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
            `,
            { ...args, user: this.props.user.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const user = data.node as GQL.IUser
                if (!user.accessTokens) {
                    throw createAggregateError(errors)
                }
                return user.accessTokens
            })
        )

    private onDidUpdateAccessToken = () => this.accessTokenUpdates.next()
}
