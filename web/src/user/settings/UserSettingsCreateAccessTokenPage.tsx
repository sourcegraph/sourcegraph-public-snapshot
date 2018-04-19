import AddIcon from '@sourcegraph/icons/lib/Add'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs/Observable'
import { concat } from 'rxjs/observable/concat'
import { catchError } from 'rxjs/operators/catchError'
import { concatMap } from 'rxjs/operators/concatMap'
import { map } from 'rxjs/operators/map'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, mutateGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { UserAreaPageProps } from '../area/UserArea'

function createAccessToken(user: GQL.ID, note: string): Observable<GQL.ICreateAccessTokenResult> {
    return mutateGraphQL(
        gql`
            mutation CreateAccessToken($user: ID!, $note: String!) {
                createAccessToken(user: $user, note: $note) {
                    id
                    token
                }
            }
        `,
        { user, note }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.createAccessToken || (errors && errors.length > 0)) {
                eventLogger.log('CreateAccessTokenFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('AccessTokenCreated')
            return data.createAccessToken
        })
    )
}

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    /** Called when a new access token is created and should be temporarily displayed to the user. */
    onDidCreateAccessToken: (result: GQL.ICreateAccessTokenResult) => void
}

interface State {
    /** The contents of the note input field. */
    note: string

    creationOrError?: 'loading' | GQL.ICreateAccessTokenResult | ErrorLike
}

/**
 * A page with a form to create an access token for a user.
 */
export class UserSettingsCreateAccessTokenPage extends React.PureComponent<Props, State> {
    public state: State = { note: '' }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(e => e.preventDefault()),
                    concatMap(() =>
                        concat(
                            [{ creationOrError: 'loading' }],
                            createAccessToken(this.props.user.id, this.state.note).pipe(
                                tap(result => {
                                    // Go back to access tokens list page and display the token secret value.
                                    this.props.onDidCreateAccessToken(result)
                                    this.props.history.push(`${this.props.match.url}/tokens`)
                                }),
                                map(result => ({ creationOrError: result })),
                                catchError(error => [{ creationOrError: asError(error) }])
                            )
                        )
                    )
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

    public render(): JSX.Element | null {
        const siteAdminViewingOtherUser =
            this.props.authenticatedUser && this.props.authenticatedUser.id !== this.props.user.id

        return (
            <div className="user-settings-create-access-token-page">
                <PageTitle title="Create access token" />
                <h2>New access token</h2>
                {siteAdminViewingOtherUser && (
                    <SiteAdminAlert className="sidebar__alert">
                        Creating access token for other user <strong>{this.props.user.username}</strong>
                    </SiteAdminAlert>
                )}
                <Form onSubmit={this.onSubmit}>
                    <div className="form-group">
                        <label htmlFor="user-settings-create-access-token-page__note">Token description</label>
                        <input
                            type="text"
                            className="form-control"
                            id="user-settings-create-access-token-page__note"
                            onChange={this.onNoteChange}
                            required={true}
                            autoFocus={true}
                            placeholder="Description"
                        />
                        <small className="form-help text-muted">What's this token for?</small>
                    </div>
                    <div className="form-group">
                        <label htmlFor="user-settings-create-access-token-page__note">Token scope</label>
                        <div className="form-check">
                            <input
                                className="form-check-input"
                                type="checkbox"
                                id="user-settings-create-access-token-page__scope"
                                checked={true}
                                disabled={true}
                            />
                            <label className="form-check-label" htmlFor="user-settings-create-access-token-page__scope">
                                <strong>all</strong> — Full control of all resources accessible to the user account
                            </label>
                        </div>
                        <small className="form-help text-muted">
                            Tokens with limited scopes are not yet supported.
                        </small>
                    </div>
                    <button
                        type="submit"
                        disabled={this.state.creationOrError === 'loading'}
                        className="btn btn-success"
                    >
                        {this.state.creationOrError === 'loading' ? (
                            <LoaderIcon className="icon-inline" />
                        ) : (
                            <AddIcon className="icon-inline" />
                        )}{' '}
                        Generate token
                    </button>
                    <Link className="btn btn-secondary ml-1" to={`${this.props.match.url}/tokens`}>
                        Cancel
                    </Link>
                </Form>
                {isErrorLike(this.state.creationOrError) && (
                    <div className="invite-form__alert alert alert-danger">
                        Error: {upperFirst(this.state.creationOrError.message)}
                    </div>
                )}
            </div>
        )
    }

    private onNoteChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setState({ note: e.currentTarget.value })

    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => this.submits.next(e)
}
