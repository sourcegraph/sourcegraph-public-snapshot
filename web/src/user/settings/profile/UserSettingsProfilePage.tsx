import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { combineLatest, concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, mergeMap, startWith, switchMap, tap } from 'rxjs/operators'
import { USER_DISPLAY_NAME_MAX_LENGTH } from '../..'
import { percentageDone } from '../../../../../shared/src/components/activation/Activation'
import { ActivationChecklist } from '../../../../../shared/src/components/activation/ActivationChecklist'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { refreshAuthenticatedUser } from '../../../auth'
import { UsernameInput } from '../../../auth/SignInSignUpCommon'
import { queryGraphQL } from '../../../backend/graphql'
import { Form } from '../../../components/Form'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserAreaRouteContext } from '../../area/UserArea'
import { UserAvatar } from '../../UserAvatar'
import { updateUser } from '../backend'
import { ErrorAlert } from '../../../components/alerts'

function queryUser(user: GQL.ID): Observable<GQL.IUser> {
    return queryGraphQL(
        gql`
            query User($user: ID!) {
                node(id: $user) {
                    ... on User {
                        id
                        username
                        displayName
                        avatarURL
                        viewerCanChangeUsername
                    }
                }
            }
        `,
        { user }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node) {
                throw createAggregateError(errors)
            }
            return data.node as GQL.IUser
        })
    )
}

interface Props extends UserAreaRouteContext, RouteComponentProps<{}> {}

interface State {
    /** The user to edit, or an error, or undefined while loading. */
    userOrError?: GQL.IUser | ErrorLike

    loading: boolean
    saved: boolean
    error?: ErrorLike

    /** undefined means unchanged from Props.user */
    username?: string
    displayName?: string
    avatarURL?: string
}

export class UserSettingsProfilePage extends React.Component<Props, State> {
    public state: State = { loading: false, saved: false }

    private componentUpdates = new Subject<Props>()
    private refreshRequests = new Subject<void>()
    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserProfile')

        const userChanges = that.componentUpdates.pipe(
            distinctUntilChanged((a, b) => a.user.id === b.user.id),
            map(({ user }) => user)
        )

        // Reset the fields upon navigation to a different user.
        that.subscriptions.add(
            userChanges.subscribe(() =>
                that.setState({
                    userOrError: undefined,
                    loading: false,
                    saved: false,
                    username: undefined,
                    displayName: undefined,
                    avatarURL: undefined,
                })
            )
        )

        // Fetch the user with all of the fields we need (Props.user might not have them all).
        that.subscriptions.add(
            combineLatest([userChanges, that.refreshRequests.pipe(startWith<void>(undefined))])
                .pipe(
                    switchMap(([user]) =>
                        queryUser(user.id).pipe(
                            catchError(error => [asError(error)]),
                            map((c): Pick<State, 'userOrError'> => ({ userOrError: c }))
                        )
                    )
                )
                .subscribe(
                    stateUpdate => that.setState(stateUpdate),
                    err => console.error(err)
                )
        )

        that.subscriptions.add(
            that.submits
                .pipe(
                    tap(event => {
                        event.preventDefault()
                        eventLogger.log('UpdateUserClicked')
                    }),
                    filter(event => event.currentTarget.checkValidity()),
                    tap(() => that.setState({ loading: true })),
                    mergeMap(event =>
                        updateUser(that.props.user.id, {
                            username: that.state.username === undefined ? null : that.state.username,
                            displayName: that.state.displayName === undefined ? null : that.state.displayName,
                            avatarURL: that.state.avatarURL === undefined ? null : that.state.avatarURL,
                        }).pipe(catchError(err => that.handleError(err)))
                    ),
                    tap(() => {
                        that.setState({ loading: false, saved: true })
                        that.props.onDidUpdateUser()

                        // Handle when username changes.
                        if (that.state.username !== undefined && that.state.username !== that.props.user.username) {
                            that.props.history.push(`/users/${that.state.username}/settings/profile`)
                            return
                        }

                        that.refreshRequests.next()
                        setTimeout(() => that.setState({ saved: false }), 500)
                    }),

                    // In case the edited user is the current user, immediately reflect the changes in the UI.
                    mergeMap(() => concat(refreshAuthenticatedUser(), [null]))
                )
                .subscribe({ error: err => that.handleError(err) })
        )
        that.componentUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-settings-profile-page">
                <PageTitle title="Profile" />
                <h2>Profile</h2>

                {that.props.activation &&
                    that.props.activation.completed &&
                    percentageDone(that.props.activation.completed) < 100 && (
                        <div className="card mb-3">
                            <div className="card-body">
                                <h3 className="mb-0">Almost there!</h3>
                                <p className="mb-0">Complete the steps below to finish onboarding to Sourcegraph.</p>
                            </div>
                            <ActivationChecklist
                                history={that.props.history}
                                steps={that.props.activation.steps}
                                completed={that.props.activation.completed}
                            />
                        </div>
                    )}

                {isErrorLike(that.state.userOrError) && <ErrorAlert error={that.state.userOrError.message} />}
                {that.state.error && <ErrorAlert error={that.state.error.message} />}
                {that.state.userOrError && !isErrorLike(that.state.userOrError) && (
                    <Form className="user-settings-profile-page__form" onSubmit={that.handleSubmit}>
                        <div className="form-group">
                            <label htmlFor="user-settings-profile-page__form-username">Username</label>
                            <UsernameInput
                                id="user-settings-profile-page__form-username"
                                className="e2e-user-settings-profile-page-username"
                                value={
                                    that.state.username === undefined
                                        ? that.state.userOrError.username
                                        : that.state.username
                                }
                                onChange={that.onUsernameFieldChange}
                                required={true}
                                disabled={
                                    !that.state.userOrError ||
                                    !that.state.userOrError.viewerCanChangeUsername ||
                                    that.state.loading
                                }
                                aria-describedby="user-settings-profile-page__form-username-help"
                            />
                            <small className="form-text text-muted">
                                A username consists of letters, numbers, hyphens (-), dots (.) and may not begin or end
                                with a dot, nor begin with a hyphen.
                            </small>
                        </div>
                        <div className="form-group">
                            <label htmlFor="user-settings-profile-page__form-display-name">Display name</label>
                            <input
                                id="user-settings-profile-page__form-display-name"
                                type="text"
                                className="form-control e2e-user-settings-profile-page__display-name"
                                value={
                                    that.state.displayName === undefined
                                        ? that.state.userOrError.displayName || ''
                                        : that.state.displayName
                                }
                                onChange={that.onDisplayNameFieldChange}
                                disabled={that.state.loading}
                                spellCheck={false}
                                placeholder="Display name"
                                maxLength={USER_DISPLAY_NAME_MAX_LENGTH}
                            />
                        </div>
                        <div className="user-settings-profile-page__avatar-row">
                            <div className="form-group user-settings-profile-page__field-column">
                                <label htmlFor="user-settings-profile-page__form-avatar-url">Avatar URL</label>
                                <input
                                    id="user-settings-profile-page__form-avatar-url"
                                    type="url"
                                    className="form-control e2e-user-settings-profile-page__avatar_url"
                                    value={
                                        that.state.avatarURL === undefined
                                            ? that.state.userOrError.avatarURL || ''
                                            : that.state.avatarURL
                                    }
                                    onChange={that.onAvatarURLFieldChange}
                                    disabled={that.state.loading}
                                    spellCheck={false}
                                    placeholder="URL to avatar photo"
                                />
                            </div>
                            {that.state.userOrError.avatarURL && (
                                <div className="user-settings-profile-page__avatar-column">
                                    <UserAvatar user={that.state.userOrError} />
                                </div>
                            )}
                        </div>
                        <button
                            className="btn btn-primary user-settings-profile-page__button e2e-user-settings-profile-page-update-profile"
                            type="submit"
                            disabled={that.state.loading}
                        >
                            Update profile
                        </button>
                        {that.state.loading && (
                            <div>
                                <LoadingSpinner className="icon-inline" />
                            </div>
                        )}
                        {that.state.saved && (
                            <p className="alert alert-success user-settings-profile-page__alert e2e-user-settings-profile-page-alert-success">
                                Profile saved!
                            </p>
                        )}
                        {window.context.sourcegraphDotComMode && (
                            <p className="mt-4">
                                <a href="https://about.sourcegraph.com/contact">Contact support</a> to delete your
                                account.
                            </p>
                        )}
                    </Form>
                )}
            </div>
        )
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ username: e.target.value })
    }

    private onDisplayNameFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ displayName: e.target.value })
    }

    private onAvatarURLFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ avatarURL: e.target.value })
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        that.submits.next(event)
    }

    private handleError = (err: Error): [] => {
        console.error(err)
        that.setState({ loading: false, saved: false, error: err })
        return []
    }
}
