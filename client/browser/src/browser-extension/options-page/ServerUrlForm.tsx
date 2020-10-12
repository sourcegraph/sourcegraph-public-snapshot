import { upperFirst } from 'lodash'
import * as React from 'react'
import { merge, Subject, Subscription } from 'rxjs'
import { debounceTime, takeUntil } from 'rxjs/operators'

export enum ConnectionErrors {
    AuthError,
    UnableToConnect,
}

const statusClassNames = {
    connecting: 'warning',
    connected: 'success',
    error: 'danger',
}

/**
 * This is the [Word-Joiner](https://en.wikipedia.org/wiki/Word_joiner) character.
 * We are using this as a &nbsp; that has no width to maintain line height when the
 * url is being updated (therefore no text is in the status indicator).
 */
const zeroWidthNbsp = '\u2060'

export interface ServerUrlFormProps {
    className?: string
    status: keyof typeof statusClassNames
    connectionError?: ConnectionErrors

    value: string
    onChange: (value: string) => void
    onSubmit: () => void
    urlHasPermissions: boolean
    requestPermissions: (url: string) => void

    /**
     * Overrides `this.props.status` and `this.state.isUpdating` in order to
     * display the `isUpdating` UI state. This is only intended for use in storybooks.
     */
    overrideUpdatingState?: boolean
}

interface State {
    isUpdating: boolean
}

export class ServerUrlForm extends React.Component<ServerUrlFormProps> {
    public state: State = { isUpdating: false }

    private inputElement = React.createRef<HTMLInputElement>()

    private componentUpdates = new Subject<State>()
    private changes = new Subject<string>()
    private submits = new Subject<void>()

    private subscriptions = new Subscription()

    constructor(props: ServerUrlFormProps) {
        super(props)

        this.subscriptions.add(
            this.changes.subscribe(value => {
                this.props.onChange(value)
                this.setState({ isUpdating: true })
            })
        )

        const submitAfterInactivity = this.changes.pipe(debounceTime(5000), takeUntil(this.submits))

        this.subscriptions.add(
            merge(this.submits, submitAfterInactivity).subscribe(() => {
                this.props.onSubmit()
                this.setState({ isUpdating: false })
            })
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.state)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            // eslint-disable-next-line react/forbid-elements
            <form className={`server-url-form ${this.props.className || ''}`} onSubmit={this.handleSubmit}>
                <label htmlFor="sourcegraph-url">Sourcegraph URL</label>
                <div className="input-group">
                    <div className="input-group-prepend">
                        <span className="input-group-text">
                            <span>
                                <span
                                    className={
                                        'server-url-form__status-indicator ' +
                                        'bg-' +
                                        (this.isUpdating ? 'secondary' : statusClassNames[this.props.status])
                                    }
                                />{' '}
                                <span className="test-connection-status">
                                    {this.isUpdating ? zeroWidthNbsp : upperFirst(this.props.status)}
                                </span>
                            </span>
                        </span>
                    </div>
                    <input
                        type="text"
                        className="form-control test-sourcegraph-url"
                        id="sourcegraph-url"
                        ref={this.inputElement}
                        value={this.props.value}
                        onChange={this.handleChange}
                        spellCheck={false}
                        autoCapitalize="off"
                        autoCorrect="off"
                    />
                </div>
                {!this.state.isUpdating && this.props.connectionError === ConnectionErrors.AuthError && (
                    <div className="alert alert-danger mt-2 mb-0">
                        Authentication to Sourcegraph failed.{' '}
                        <a href={this.props.value} target="_blank" rel="noopener noreferrer">
                            Sign in to your instance
                        </a>{' '}
                        to continue.
                    </div>
                )}
                {!this.state.isUpdating && this.props.connectionError === ConnectionErrors.UnableToConnect && (
                    <div className="alert alert-danger mt-2 mb-0">
                        <p>
                            Unable to connect to{' '}
                            <a href={this.props.value} target="_blank" rel="noopener noreferrer">
                                {this.props.value}
                            </a>
                            . Ensure the URL is correct and you are{' '}
                            <a href={this.props.value + '/sign-in'} target="_blank" rel="noopener noreferrer">
                                signed in
                            </a>
                            .
                        </p>
                        {!this.props.urlHasPermissions && (
                            <p>
                                You may need to{' '}
                                <a href="#" onClick={this.requestServerURLPermissions}>
                                    grant the Sourcegraph browser extension additional permissions
                                </a>{' '}
                                for this URL.
                            </p>
                        )}
                        <p className="mb-0">
                            <b>Site admins:</b> ensure that{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/config/site_config"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                all users can create access tokens
                            </a>
                            .
                        </p>
                    </div>
                )}
            </form>
        )
    }

    private handleChange = ({ target: { value } }: React.ChangeEvent<HTMLInputElement>): void => {
        this.changes.next(value)
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()

        this.submits.next()
    }

    private requestServerURLPermissions = (): void => this.props.requestPermissions(this.props.value)

    private get isUpdating(): boolean {
        if (typeof this.props.overrideUpdatingState !== 'undefined') {
            console.warn(
                '<ServerUrlForm /> - You are using the `overrideUpdatingState` prop which is ' +
                    'only intended for use with storybooks. Keeping this state in multiple places can ' +
                    'lead to race conditions and will be hard to maintain.'
            )

            return this.props.overrideUpdatingState
        }

        return this.state.isUpdating
    }
}
