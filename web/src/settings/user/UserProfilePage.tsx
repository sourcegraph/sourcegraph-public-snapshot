import { Loader } from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/mergeMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser, fetchCurrentUser } from '../../auth'
import { PageTitle } from '../../components/PageTitle'
import { events } from '../../tracking/events'
import { createUser, updateUser } from '../backend'
import { VALID_USERNAME_REGEXP } from '../validation'
import { UserAvatar } from './UserAvatar'

interface Props {
    location: H.Location
    history: H.History
}

interface State {
    user: GQL.IUser | null
    error?: Error
    loading?: boolean
    saved?: boolean
    username: string
    displayName: string
}

/**
 * A landing page for the user to sign in or register, if not authed
 */
export class UserProfilePage extends React.Component<Props, State> {
    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const redirectError = this.getRedirectError()
        this.state = {
            user: null,
            username: '',
            displayName: '',
            error: redirectError ? new Error(redirectError) : undefined,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(
            user => this.setState({ user, username: (user && user.username) || '', displayName: (user && user.displayName) || '' }),
            error => this.setState({ error })
        ))
        this.subscriptions.add(
            this.submits
                .do(event => {
                    event.preventDefault()
                    events.UpdateUserClicked.log()
                })
                .filter(event => event.currentTarget.checkValidity())
                .do(() => this.setState({ loading: true }))
                .mergeMap(event =>
                    this.requireBackfill() ?
                        createUser({ username: this.state.username, displayName: this.state.displayName || this.state.username })
                            .do(() => window.context.requireUserBackfill = false)
                            .catch(this.handleError) :
                        updateUser({ username: this.state.username, displayName: this.state.displayName })
                            .catch(this.handleError)
                )
                .do(() => this.setState({ loading: false, error: undefined, saved: true }))
                .mergeMap(() => fetchCurrentUser().concat([null]))
                .subscribe(
                    () => {
                        const searchParams = new URLSearchParams(this.props.location.search)
                        const returnTo = searchParams.get('returnTo')
                        if (returnTo) {
                            const newURL = new URL(returnTo, window.location.href)
                            this.props.history.replace(newURL.pathname + newURL.search + newURL.hash)
                        } else {
                            // just take back to settings
                            this.props.history.replace('/settings')
                        }
                    },
                    this.handleError
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className='user-profile-page'>
                <PageTitle title='Profile' />
                <h1>Your profile</h1>
                {this.requireBackfill() && <p className='alert alert-danger'>Please complete your profile to continue using Sourcegraph.</p>}
                {this.state.error && <p className='alert alert-danger'>{upperFirst(this.state.error.message)}</p>}
                {this.state.saved && <p className='alert alert-success'>Profile saved!</p>}
                <form className='user-profile-page__form' onSubmit={this.handleSubmit}>
                    <div className='user-profile-page__avatar-row'>
                        <div className='user-profile-page__avatar-column'>
                            <UserAvatar />
                        </div>
                        <div className='form-group'>
                            <label>Username</label>
                            <input
                                type='text'
                                className='ui-text-box'
                                value={this.state.username}
                                onChange={this.onUsernameFieldChange}
                                pattern={VALID_USERNAME_REGEXP.toString().slice(1, -1)}
                                required={true}
                                disabled={this.state.loading}
                                spellCheck={false}
                                placeholder='Username'
                            />
                            <small className='form-text'>A username consists of letters, numbers, hyphens (-) and may not begin or end with a hyphen</small>
                        </div>
                    </div>
                    <div className='form-group'>
                        <label>Email</label>
                        <input
                            readOnly={true}
                            type='email'
                            className='ui-text-box'
                            value={this.state.user && this.state.user.email || ''}
                            disabled={true}
                            spellCheck={false}
                            placeholder='Email'
                        />
                    </div>
                    <div className='form-group'>
                        <label>Display name (optional)</label>
                        <input
                            type='text'
                            className='ui-text-box'
                            value={this.state.displayName}
                            onChange={this.onDisplayNameFieldChange}
                            disabled={this.state.loading}
                            spellCheck={false}
                            placeholder='Display name'
                        />
                    </div>
                    <button className='btn btn-primary user-profile-page__button' type='submit' disabled={this.state.loading}>
                        Update profile
                    </button>
                    {this.state.loading && <div className='icon-inline'><Loader className='icon-inline'/></div>}
                </form>
            </div>
        )
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ username: e.target.value })
    }

    private onDisplayNameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ displayName: e.target.value })
    }

    private requireBackfill(): boolean {
        return new URLSearchParams(this.props.location.search).get('backfill') !== null
    }

    private getRedirectError(): string | undefined {
        const code = new URLSearchParams(this.props.location.search).get('error')
        if (!code) {
            return undefined
        }
        switch (code) {
            case 'err_username_exists':
                return 'The username you selected is already taken, please try again.'
            case 'err_email_exists':
                return 'The email you selected is already taken, please try again.'
        }
        return 'There was an error creating your profile, please try again.'
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        this.submits.next(event)
    }

    private handleError = (err: Error) => {
        console.error(err)
        this.setState({ loading: false, saved: false, error: err })
        return []
    }
}
