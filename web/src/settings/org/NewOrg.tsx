import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { catchError } from 'rxjs/operators/catchError'
import { filter } from 'rxjs/operators/filter'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { createOrg } from '../backend'
import { VALID_ORG_NAME_REGEXP } from '../validation'

export interface Props {
    history: H.History
}

export interface State {
    /**
     * Current value of the organization name input
     */
    name: string

    displayName: string

    /**
     * Holds any error returned by the remote GraphQL endpoint on failed requests.
     */
    error?: Error

    /**
     * True if the form is currently being submitted
     */
    loading: boolean
}

export class NewOrg extends React.Component<Props, State> {
    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            loading: false,
            name: '',
            displayName: '',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('NewOrg')
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(event => {
                        event.preventDefault()
                        eventLogger.log('CreateNewOrgClicked')
                    }),
                    filter(event => event.currentTarget.checkValidity()),
                    mergeMap(event =>
                        createOrg(this.state).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    org => {
                        this.props.history.push(`/settings/orgs/${org.name}`)
                    },
                    error => {
                        console.error(error)
                    }
                )
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="new-organization">
                <PageTitle title="New organization" />
                <form className="settings-form" onSubmit={this.onSubmit}>
                    <h1>Create a new organization</h1>

                    <p>
                        Create a Sourcegraph organization to bring the discussion to the code. Learn more about{' '}
                        <a href="https://about.sourcegraph.com/products/editor" target="_blank">
                            collaboration in Sourcegraph
                        </a>.
                    </p>
                    {this.state.error && <p className="form-text text-error">{upperFirst(this.state.error.message)}</p>}
                    <div className="form-group">
                        <label>Organization name</label>
                        <input
                            type="text"
                            className="form-control"
                            placeholder="acme-corp"
                            pattern={VALID_ORG_NAME_REGEXP.toString().slice(1, -1)}
                            required={true}
                            autoCorrect="off"
                            autoComplete="off"
                            autoFocus={true}
                            value={this.state.name}
                            onChange={this.onNameChange}
                            disabled={this.state.loading}
                        />
                        <small className="form-text">
                            An organization name consists of letters, numbers, hyphens (-) and may not begin or end with
                            a hyphen
                        </small>
                    </div>

                    <div className="form-group">
                        <label>Display name</label>
                        <input
                            type="text"
                            className="form-control"
                            placeholder="ACME Corporation"
                            required={true}
                            autoCorrect="off"
                            value={this.state.displayName}
                            onChange={this.onDisplayNameChange}
                            disabled={this.state.loading}
                        />
                    </div>

                    <button type="submit" className="btn btn-primary" disabled={this.state.loading}>
                        Create organization
                    </button>
                    {this.state.loading && <LoaderIcon className="icon-inline" />}
                </form>
            </div>
        )
    }

    private onNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        const hyphenatedName = event.currentTarget.value.replace(/\s/g, '-')
        this.setState({ name: hyphenatedName })
    }

    private onDisplayNameChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.setState({ displayName: event.currentTarget.value })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        this.submits.next(event)
    }
}
