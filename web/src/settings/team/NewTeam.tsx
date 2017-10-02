import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/distinctUntilChanged'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser, fetchCurrentUser } from '../../auth'
import { events } from '../../tracking/events'
import { createOrg } from '../backend'
import { VALID_ORG_NAME_REGEXP, VALID_USERNAME_REGEXP } from '../validation'

export interface Props {
    history: H.History
}

export interface State {

    /**
     * Current value of the team name input
     */
    name: string

    username: string

    displayName: string

    email: string

    /**
     * Holds any error returned by the remote GraphQL endpoint on failed requests.
     */
    error?: Error

    /**
     * True if the form is currently being submitted
     */
    loading: boolean
}

export class NewTeam extends React.Component<Props, State> {

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    constructor() {
        super()
        this.state = {
            loading: false,
            displayName: '',
            name: '',
            username: '',
            email: ''
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser
                .filter((user: GQL.IUser | null): user is GQL.IUser => !!user)
                .distinctUntilChanged((a, b) => !!(a && b && a.id === b.id))
                .map(({ email }) => {
                    email = email || ''
                    let username = ''
                    if (email) {
                        username = email.split('@')[0]
                        if (!VALID_USERNAME_REGEXP.test(username)) {
                            username = ''
                        }
                    }
                    return { email, username }
                })
                .subscribe(state => this.setState(state))
        )
        this.subscriptions.add(
            this.submits
                .do(event => {
                    event.preventDefault()
                    events.CreateNewOrgClicked.log()
                })
                .filter(event => event.currentTarget.checkValidity())
                .mergeMap(event =>
                    createOrg(this.state)
                        .catch(error => {
                            console.error(error)
                            this.setState({ error })
                            return []
                        })
                )
                .mergeMap(team => fetchCurrentUser().concat([team]))
                .subscribe(team => {
                    this.props.history.push(`/settings/teams/${team.name}`)
                }, error => {
                    console.error(error)
                })
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className='new-team'>
                <form className='settings-form' onSubmit={this.onSubmit}>

                    <h1>Create a new team</h1>

                    <p>
                        Create a Sourcegraph team to bring the discussion to the code.
                        Learn more about <a href='https://about.sourcegraph.com/products/editor' target='_blank'>collaboration in Sourcegraph</a>.
                    </p>
                    {this.state.error && <p className='form-text text-error'>{upperFirst(this.state.error.message)}</p>}
                    <div className='form-group'>
                        <label>Team name</label>
                        <input
                            type='text'
                            className='ui-text-box'
                            placeholder='your-team'
                            pattern={VALID_ORG_NAME_REGEXP.toString().slice(1, -1)}
                            required={true}
                            autoCorrect='off'
                            autoComplete='off'
                            autoFocus={true}
                            value={this.state.name}
                            onChange={this.onNameChange}
                            disabled={this.state.loading}
                        />
                        <small className='form-text'>A team name consists of letters, numbers, hyphens (-) and may not begin or end with a hyphen</small>
                    </div>

                    <div className='form-group'>
                        <label>Your new username</label>
                        <input
                            type='text'
                            className='ui-text-box'
                            placeholder='yourusername'
                            pattern={VALID_USERNAME_REGEXP.toString().slice(1, -1)}
                            required={true}
                            autoCorrect='off'
                            value={this.state.username}
                            onChange={this.onUserNameChange}
                            disabled={this.state.loading}
                        />
                        <small className='form-text'>A username consists of letters, numbers, hyphens (-) and may not begin or end with a hyphen</small>
                    </div>

                    <div className='form-group'>
                        <label>Your display name</label>
                        <input
                            type='text'
                            className='ui-text-box'
                            placeholder='Your Name'
                            required={true}
                            autoCorrect='off'
                            value={this.state.displayName}
                            onChange={this.onDisplayNameChange}
                            disabled={this.state.loading}
                        />
                    </div>

                    <div className='form-group'>
                        <label>Your company email</label>
                        <input
                            type='email'
                            className='ui-text-box'
                            placeholder='you@yourcompany.com'
                            required={true}
                            autoCorrect='off'
                            value={this.state.email}
                            onChange={this.onEmailChange}
                            disabled={this.state.loading}
                        />
                    </div>

                    <button type='submit' className='btn btn-primary' disabled={this.state.loading}>Create Team</button>
                    {this.state.loading && <LoaderIcon className='icon-inline' />}

                </form>
            </div>
        )
    }

    private onNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const hyphenatedName = event.currentTarget.value.replace(/\s/g, '-')
        this.setState({ name: hyphenatedName })
    }

    private onUserNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.setState({ username: event.currentTarget.value })
    }

    private onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.setState({ displayName: event.currentTarget.value })
    }

    private onEmailChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.setState({ email: event.currentTarget.value })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        this.submits.next(event)
    }
}
